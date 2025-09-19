package db

import (
	"context"
	"fmt"
	_ "github.com/jackc/pgx/v5"
	"log"
	"red-courier/internal/config"
	"red-courier/internal/redis"
	sqlbuilder "red-courier/internal/sql_builder"
	"red-courier/internal/util"
)

//TODO extract SQL generation logic to separate package

// FetchRows retrieves rows from the specified table based on the task configuration.
// It applies any static WHERE clauses and tracking filters, and returns the results as a slice of maps.
func (db *Database) FetchRows(ctx context.Context, taskCfg config.TaskConfig, redisClient *redis.RedisClient) ([]map[string]any, error) {
	cols := resolveColumns(taskCfg)
	if len(cols) == 0 {
		return nil, fmt.Errorf("no columns resolved for task: %s", taskCfg.Name)
	}

	// Resolve schema.table and tracking context
	var lastValPtr *string
	var trackingSpec *sqlbuilder.TrackingSpec
	if taskCfg.Tracking != nil {
		resolvedCol := taskCfg.ResolveColumn(taskCfg.Tracking.Column)
		trackingSpec = &sqlbuilder.TrackingSpec{
			Column:       resolvedCol,
			Operator:     taskCfg.Tracking.Operator,
			LastValueKey: taskCfg.Tracking.LastValueKey,
		}

		val, err := redisClient.Client.Get(ctx, taskCfg.Tracking.LastValueKey).Result()
		if err != nil && err.Error() != "redis: nil" && val == "" {
			return nil, fmt.Errorf("failed to fetch last value for tracking: %w", err)
		}
		if val != "" {
			lastValPtr = &val
		}
	}

	spec, _ := sqlbuilder.FromQualifiedTable(taskCfg.Table, cols, taskCfg.Where, trackingSpec, lastValPtr)
	plan, err := sqlbuilder.BuildSelect(spec)
	if err != nil {
		return nil, err
	}

	logSQL := taskCfg.EffectiveLogSQL(db.LogSql)
	if logSQL {
		// Keep it structured and readable. Redact/limit args if needed.
		//TODO make configurable to log full args
		//TODO consider using a proper SQL formatter
		//TODO consider logging to a file instead of stdout
		//TODO consider using a proper structured logger like zap or logrus
		redacted := make([]any, len(plan.Args))
		for i, a := range plan.Args {
			s := fmt.Sprint(a)
			if len(s) > 256 {
				s = s[:256] + "…(truncated)"
			}
			redacted[i] = s
		}
		log.Printf("[task:%s] SQL: %s  ARGS: %v", taskCfg.Name, plan.SQL, redacted)
	}

	rows, err := db.Pool.Query(ctx, plan.SQL, plan.Args...)
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
		rowMap := make(map[string]any, len(values))
		for i, fd := range rows.FieldDescriptions() {
			rowMap[string(fd.Name)] = values[i]
		}
		results = append(results, rowMap)

		// compute first-run checkpoint if needed
		if plan.FirstRun && plan.TrackingCol != "" {
			v := rowMap[plan.TrackingCol]
			if maxVal == nil || util.CompareAny(v, maxVal) > 0 {
				maxVal = v
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Persist initial checkpoint if first run
	if plan.FirstRun && maxVal != nil && taskCfg.Tracking != nil {
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
