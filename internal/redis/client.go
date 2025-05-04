package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisClient(cfg RedisConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisClient{Client: client}
}

func (r *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisClient) SetString(ctx context.Context, key, value string) error {
	return r.Client.Set(ctx, key, value, 0).Err()
}

func (r *RedisClient) PushToList(ctx context.Context, key string, value any) error {
	return r.Client.LPush(ctx, key, value).Err()
}

func (r *RedisClient) AddToSet(ctx context.Context, key string, value any) error {
	return r.Client.SAdd(ctx, key, value).Err()
}

func (r *RedisClient) HSetField(ctx context.Context, key string, field, value any) error {
	return r.Client.HSet(ctx, key, field, value).Err()
}

func (r *RedisClient) AddToSortedSet(ctx context.Context, key string, score float64, value any) error {
	return r.Client.ZAdd(ctx, key, redis.Z{Score: score, Member: value}).Err()
}

func (r *RedisClient) AddToStream(ctx context.Context, key string, values map[string]any) error {
	return r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: values,
	}).Err()
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
