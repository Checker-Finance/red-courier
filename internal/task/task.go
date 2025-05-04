package task

import (
	"context"
	"fmt"
	"log"

	"red-courier/internal/config"
	"red-courier/internal/db"
	"red-courier/internal/redis"
	"red-courier/internal/redis/loaders"
)

type Task struct {
	Config      config.TaskConfig
	DB          *db.Database
	RedisClient *redis.RedisClient
	Loader      loaders.RedisLoader
}

func NewTask(cfg config.TaskConfig, db *db.Database, rdb *redis.RedisClient) (*Task, error) {
	loader, err := loaders.NewRedisLoader(cfg)
	if err != nil {
		return nil, fmt.Errorf("loader error: %w", err)
	}

	return &Task{
		Config:      cfg,
		DB:          db,
		RedisClient: rdb,
		Loader:      loader,
	}, nil
}

func (t *Task) Run(ctx context.Context) error {
	log.Printf("Running task for table: %s (alias: %s)", t.Config.Table, t.Config.Alias)

	rows, err := t.DB.FetchRows(ctx, t.Config)
	if err != nil {
		return fmt.Errorf("error fetching rows: %w", err)
	}

	keyBase := t.Config.Alias
	if keyBase == "" {
		keyBase = t.Config.Table
	}

	err = t.Loader.Load(ctx, t.RedisClient.Client, keyBase, rows)
	if err != nil {
		return fmt.Errorf("error writing to Redis: %w", err)
	}

	log.Printf("Task complete: %d rows written to Redis", len(rows))
	return nil
}
