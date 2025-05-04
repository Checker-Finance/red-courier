package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"red-courier/internal/config"
	"red-courier/internal/db"
	"red-courier/internal/redis"
	"red-courier/internal/scheduler"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pg, err := db.NewDatabase(cfg.Postgres)
	if err != nil {
		log.Fatalf("Postgres error: %v", err)
	}
	defer pg.Close()

	rdb := redis.NewRedisClient(redis.RedisConfig(cfg.Redis))
	if err != nil {
		log.Fatalf("Redis error: %v", err)
	}
	defer rdb.Close()

	ctx := context.Background()
	sched, err := scheduler.NewScheduler(ctx, cfg, pg, rdb)
	if err != nil {
		log.Fatalf("Scheduler setup failed: %v", err)
	}

	go sched.Start()

	// Handle graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	sched.Stop()
}
