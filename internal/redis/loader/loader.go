package loader

import (
	"context"
	"fmt"
	
	"red-courier/internal/config"
	"red-courier/internal/redis"
)

type Loader interface {
	Load(ctx context.Context, rows []map[string]any, cfg config.TaskConfig, r *redis.RedisClient) error
}

func NewLoader(cfg config.TaskConfig) (Loader, error) {
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
