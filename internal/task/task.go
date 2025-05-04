// internal/task/task.go
package task

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"red-courier/internal/config"
	"red-courier/internal/db"
	"red-courier/internal/redis"
	"red-courier/internal/redis/loader"
)

type Task struct {
	Config      config.TaskConfig
	DB          *db.Database
	RedisClient *redis.RedisClient
	Loader      loader.Loader
}

func NewTask(cfg config.TaskConfig, dbConn *db.Database, redisConn *redis.RedisClient) (*Task, error) {
	ld, err := loader.NewLoader(cfg)
	if err != nil {
		return nil, err
	}
	return &Task{
		Config:      cfg,
		DB:          dbConn,
		RedisClient: redisConn,
		Loader:      ld,
	}, nil
}

func (t *Task) Run(ctx context.Context) error {
	log.Printf("[task:%s] Running task", t.Config.Name)

	rows, err := t.DB.FetchRows(ctx, t.Config, t.RedisClient)
	if err != nil {
		return fmt.Errorf("failed to fetch rows: %w", err)
	}

	if err := t.Loader.Load(ctx, rows, t.Config, t.RedisClient); err != nil {
		return fmt.Errorf("failed to load into Redis: %w", err)
	}

	if t.Config.Tracking != nil && len(rows) > 0 {
		trackingCol := t.Config.ResolveColumn(t.Config.Tracking.Column)
		var maxVal any
		for _, row := range rows {
			v := row[trackingCol]
			if maxVal == nil || compareAny(v, maxVal) > 0 {
				maxVal = v
			}
		}
		if maxStr, ok := toRedisString(maxVal); ok {
			if err := t.RedisClient.SetString(ctx, t.Config.Tracking.LastValueKey, maxStr); err != nil {
				log.Printf("[task:%s] Failed to persist tracking value: %v", t.Config.Name, err)
			} else {
				log.Printf("[task:%s] Updated checkpoint: %s = %s", t.Config.Name, t.Config.Tracking.LastValueKey, maxStr)
			}
		}
	}

	log.Printf("[task:%s] Completed with %d rows", t.Config.Name, len(rows))
	return nil
}

func compareAny(a, b any) int {
	switch a := a.(type) {
	case int64:
		return compareInts(a, b.(int64))
	case float64:
		return compareFloats(a, b.(float64))
	case string:
		return compareStrings(a, b.(string))
	case time.Time:
		return compareTimes(a, b.(time.Time))
	default:
		return 0
	}
}

func compareInts(a, b int64) int {
	if a > b {
		return 1
	} else if a < b {
		return -1
	} else {
		return 0
	}
}
func compareFloats(a, b float64) int {
	if a > b {
		return 1
	} else if a < b {
		return -1
	} else {
		return 0
	}
}
func compareStrings(a, b string) int { return strings.Compare(a, b) }
func compareTimes(a, b time.Time) int {
	if a.After(b) {
		return 1
	} else if a.Before(b) {
		return -1
	}
	return 0
}

func toRedisString(v any) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case int64, float64:
		return fmt.Sprintf("%v", val), true
	case time.Time:
		return val.Format(time.RFC3339Nano), true
	default:
		return "", false
	}
}
