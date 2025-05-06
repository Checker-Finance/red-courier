package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5"
	"red-courier/internal/config"
	"red-courier/internal/redis"
	"red-courier/internal/util"
)

func (db *Database) FetchRows(ctx context.Context, taskCfg config.TaskConfig, redisClient *redis.RedisClient) ([]map[string]any, error) {
	table := taskCfg.Table
	columns := resolveColumns(taskCfg)
	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns resolved for task: %s", taskCfg.Name)
	}

	schema, table := "public", taskCfg.Table
	if strings.Contains(taskCfg.Table, ".") {
		parts := strings.SplitN(taskCfg.Table, ".", 2)
		schema, table = parts[0], parts[1]
	}

	query := fmt.Sprintf(`SELECT %s FROM "%s"."%s"`, strings.Join(columns, ", "), schema, table)
	var args []any
	var firstRun bool
	var trackingCol string

	// Apply tracking filter if present
	if taskCfg.Tracking != nil {
		trackingCol = taskCfg.ResolveColumn(taskCfg.Tracking.Column)
		lastValue, err := redisClient.Client.Get(ctx, taskCfg.Tracking.LastValueKey).Result()
		if lastValue == "" {
			log.Printf("[task:%s] No checkpoint found in Redis. Fetching all rows.", taskCfg.Name)
			firstRun = true
		} else if err != nil {
			return nil, fmt.Errorf("failed to fetch last value for tracking: %w", err)
		} else if lastValue != "" {
			query += fmt.Sprintf(" WHERE %s %s $1", trackingCol, taskCfg.Tracking.Operator)
			args = append(args, lastValue)
		}
	}

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]any
	var maxVal any

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row values: %w", err)
		}

		rowMap := make(map[string]any)
		for i, field := range rows.FieldDescriptions() {
			rowMap[string(field.Name)] = values[i]
		}

		results = append(results, rowMap)

		// Track max value if full fetch and tracking is enabled
		if firstRun && trackingCol != "" {
			v := rowMap[trackingCol]
			if maxVal == nil || util.CompareAny(v, maxVal) > 0 {
				maxVal = v
			}
		}
	}

	// On first run, store the new checkpoint
	if firstRun && maxVal != nil {
		if maxStr, ok := util.ToRedisString(maxVal); ok {
			if err := redisClient.SetString(ctx, taskCfg.Tracking.LastValueKey, maxStr); err != nil {
				log.Printf("[task:%s] Failed to persist initial checkpoint: %v", taskCfg.Name, err)
			} else {
				log.Printf("[task:%s] Stored initial checkpoint: %s = %s", taskCfg.Name, taskCfg.Tracking.LastValueKey, maxStr)
			}
		}
	}

	return results, nil
}

func resolveColumns(taskCfg config.TaskConfig) []string {
	var logicalCols []string
	switch taskCfg.Structure {
	case "stream":
		logicalCols = taskCfg.Fields
	default:
		logicalCols = []string{taskCfg.Key, taskCfg.Value, taskCfg.Score}
	}

	// Include tracking column
	if taskCfg.Tracking != nil && taskCfg.Tracking.Column != "" {
		logicalCols = append(logicalCols, taskCfg.Tracking.Column)
	}

	unique := make(map[string]struct{})
	var resolved []string
	for _, logical := range logicalCols {
		if logical == "" {
			continue
		}
		actual := taskCfg.ResolveColumn(logical)
		if _, seen := unique[actual]; !seen {
			unique[actual] = struct{}{}
			resolved = append(resolved, actual)
		}
	}
	return resolved
}
