package loader

import (
	"context"
	"fmt"
	"red-courier/internal/config"
	"red-courier/internal/redis"
)

type MapLoader struct {
	KeyField   string
	ValueField string
}

func (l *MapLoader) Load(ctx context.Context, rows []map[string]any, cfg config.TaskConfig, r *redis.RedisClient) error {
	key := cfg.EffectiveRedisKey()
	for _, row := range rows {
		k, kOk := row[cfg.ResolveColumn(cfg.Key)]
		v, vOk := row[cfg.ResolveColumn(cfg.Value)]
		if !kOk || !vOk {
			continue
		}
		if err := r.Client.HSet(ctx, key, k, v).Err(); err != nil {
			return fmt.Errorf("failed to HSET to Redis map: %w", err)
		}
	}
	return nil
}
