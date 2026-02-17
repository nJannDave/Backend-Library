package config

import (
	"stmnplibrary/log"

	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

func ProviderCTX() context.Context {
	return context.TODO()
}

func ConnectRedis(ctx context.Context) *redis.Client {
	rds := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if err := rds.Ping(ctx).Err(); err != nil {
		log.LogConfig("failed connect to redis", "connect_redis", err)
		return nil
	}
	return rds
}
