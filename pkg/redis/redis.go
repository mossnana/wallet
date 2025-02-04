package redis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func EnvClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}
