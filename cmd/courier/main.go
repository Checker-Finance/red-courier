package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"red-courier/internal/config"
	"red-courier/internal/db"
	"red-courier/internal/redis"
	"red-courier/internal/scheduler"
)

func main() {
	defaultPath := "config.yaml"
	envPath := os.Getenv("RED_COURIER_CONFIG")
	if envPath != "" {
		defaultPath = envPath
	}

	cfgPath := flag.String("config", defaultPath, "path to the config file (YAML)")
	flag.Parse()

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pg, err := db.NewDatabase(*cfg)
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
	port := cfg.Server.Port
	log.Printf("Server starting on port %s", port)

	if port == "" {
		port = ":8080"
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		_ = http.ListenAndServe(port, mux)
	}()

	// Handle graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	sched.Stop()
}
