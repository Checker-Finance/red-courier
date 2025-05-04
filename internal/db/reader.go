package db

import (
	"context"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5"
	"red-courier/internal/config"
	"red-courier/internal/redis"
)

func (db *Database) FetchRows(ctx context.Context, taskCfg config.TaskConfig, redisClient *redis.RedisClient) ([]map[string]any, error) {
	table := taskCfg.Table
	columns := resolveColumns(taskCfg)
	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns resolved for task: %s", taskCfg.Name)
	}

	query := fmt.Sprintf(`SELECT %s FROM %q`, strings.Join(columns, ", "), table)
	var args []any

	// Apply tracking filter if present
	if taskCfg.Tracking != nil {
		column := taskCfg.ResolveColumn(taskCfg.Tracking.Column)
		lastValue, err := redisClient.Client.Get(ctx, taskCfg.Tracking.LastValueKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch last value for tracking: %w", err)
		}
		if lastValue != "" {
			query += fmt.Sprintf(" WHERE %s %s $1", column, taskCfg.Tracking.Operator)
			args = append(args, lastValue)
		}
	}

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []map[string]any

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
