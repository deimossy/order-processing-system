package redis

import (
	"context"
	"github.com/deimossy/order-processing-system/internal/user/config"
	"github.com/redis/go-redis/v9"
	"log"
)

func NewRedisClient(ctx context.Context, cfg config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
	})

	timeout, cancel := context.WithTimeout(ctx, cfg.RedisPingTimeout)
	defer cancel()

	if err := rdb.Ping(timeout).Err(); err != nil {
		log.Fatal(err)
	}

	return rdb
}
