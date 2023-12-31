package testutils

import (
	"os"

	redis "github.com/redis/go-redis/v9"
)

func RedisClient() *redis.Client {
	redisURL := "localhost:6379"
	if os.Getenv("GACR_REDIS_URL") != "" {
		redisURL = os.Getenv("GACR_REDIS_URL")
	}
	return redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	})
}
