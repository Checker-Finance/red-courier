// internal/db/reader.go
package db

import (
	"context"
	"fmt"
	"strings"

	"red-courier/internal/config"
)

func (db *Database) FetchRows(ctx context.Context, taskCfg config.TaskConfig) ([]map[string]any, error) {
	table := taskCfg.Table

	columns := resolveColumns(taskCfg)
	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns resolved for task: %s", taskCfg.Name)
	}

	query := fmt.Sprintf(`SELECT %s FROM %q`, strings.Join(columns, ", "), table)

	rows, err := db.Pool.Query(ctx, query)
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
