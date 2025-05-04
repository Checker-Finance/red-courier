package loaders

import (
	"context"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"red-courier/internal/config"
)

type RedisLoader interface {
	Load(ctx context.Context, client *goredis.Client, key string, rows []map[string]any) error
}

func NewRedisLoader(cfg config.TaskConfig) (RedisLoader, error) {
	switch cfg.Structure {
	case "map":
		return &MapLoader{
			KeyField:   cfg.Key,
			ValueField: cfg.Value,
		}, nil

	case "list":
		return &ListLoader{
			ValueField: cfg.Value,
		}, nil

	case "set":
		return &SetLoader{
			ValueField: cfg.Value,
		}, nil

	case "sorted_set":
		return &SortedSetLoader{
			ValueField: cfg.Value,
			ScoreField: cfg.Score,
		}, nil

	case "stream":
		return &StreamLoader{
			Fields: cfg.Fields,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported Redis structure: %s", cfg.Structure)
	}
}
