package loader

import (
	"context"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"red-courier/internal/config"
	"red-courier/internal/redis"
)

type StreamLoader struct {
	Fields []string
}

func (l *StreamLoader) Load(ctx context.Context, rows []map[string]any, cfg config.TaskConfig, r *redis.RedisClient) error {
	key := cfg.EffectiveRedisKey()
	for _, row := range rows {
		fields := make(map[string]any)
		for _, logical := range cfg.Fields {
			col := cfg.ResolveColumn(logical)
			val, ok := row[col]
			if !ok {
				continue
			}
			fields[logical] = val
		}

		if len(fields) == 0 {
			continue
		}

		args := &goredis.XAddArgs{
			Stream: key,
			Values: fields,
		}
		if err := r.Client.XAdd(ctx, args).Err(); err != nil {
			return fmt.Errorf("failed to XADD to Redis stream: %w", err)
		}
	}
	return nil
}
