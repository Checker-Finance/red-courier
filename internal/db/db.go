package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"red-courier/internal/config"
)

type Database struct {
	Pool   *pgxpool.Pool
	LogSql bool
}

func NewDatabase(cfg config.Config) (*Database, error) {
	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s pool_max_conns=10",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName, cfg.Postgres.SSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Ping to verify the connection is valid
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	return &Database{Pool: pool, LogSql: cfg.LogSQL}, nil
}

func (db *Database) Close() {
	db.Pool.Close()
}
