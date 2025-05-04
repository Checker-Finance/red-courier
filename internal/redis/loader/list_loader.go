package loader

import (
	"context"
	"fmt"
	"red-courier/internal/config"
	"red-courier/internal/redis"
)

type ListLoader struct {
	ValueField string
}

func (l *ListLoader) Load(ctx context.Context, rows []map[string]any, cfg config.TaskConfig, r *redis.RedisClient) error {
	key := cfg.EffectiveRedisKey()
	for _, row := range rows {
		val, ok := row[cfg.ResolveColumn(cfg.Value)]
		if !ok {
			continue
		}
		if err := r.Client.LPush(ctx, key, val).Err(); err != nil {
			return fmt.Errorf("failed to LPUSH to Redis list: %w", err)
		}
	}
	return nil
}
