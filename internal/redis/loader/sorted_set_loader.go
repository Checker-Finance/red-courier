package loader

import (
	"context"
	"fmt"
	"red-courier/internal/config"
	"red-courier/internal/redis"
	"strconv"
)

type SortedSetLoader struct {
	ValueField string
	ScoreField string
}

func (l *SortedSetLoader) Load(ctx context.Context, rows []map[string]any, cfg config.TaskConfig, r *redis.RedisClient) error {
	key := cfg.EffectiveRedisKey()
	for _, row := range rows {
		val, valOk := row[cfg.ResolveColumn(cfg.Value)]
		scoreRaw, scoreOk := row[cfg.ResolveColumn(cfg.Score)]
		if !valOk || !scoreOk {
			continue
		}

		var score float64
		switch s := scoreRaw.(type) {
		case float64:
			score = s
		case int64:
			score = float64(s)
		case string:
			parsed, err := strconv.ParseFloat(s, 64)
			if err != nil {
				continue
			}
			score = parsed
		default:
			continue
		}

		if err := r.AddToSortedSet(ctx, key, score, val); err != nil {
			return fmt.Errorf("failed to ZADD to Redis sorted set: %w", err)
		}
	}
	return nil
}
