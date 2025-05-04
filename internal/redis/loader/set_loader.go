package loader

import (
	"context"
	"fmt"
	"red-courier/internal/config"
	"red-courier/internal/redis"
)

type SetLoader struct {
	ValueField string
}

func (l *SetLoader) Load(ctx context.Context, rows []map[string]any, cfg config.TaskConfig, r *redis.RedisClient) error {
	key := cfg.EffectiveRedisKey()
	for _, row := range rows {
		val, ok := row[cfg.ResolveColumn(cfg.Value)]
		if !ok {
			continue
		}
		if err := r.Client.SAdd(ctx, key, val).Err(); err != nil {
			return fmt.Errorf("failed to SADD to Redis set: %w", err)
		}
	}
	return nil
}
