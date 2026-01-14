package redis

import (
	"strconv"

	"github.com/ayuspoudel/sentinel-sre/pkg/env"
	"github.com/redis/go-redis/v9"
)

func New() *redis.Client {
	url := env.MustEnv("REDIS_URL", false)
	password := env.MustEnv("REDIS_PASSWORD", true)
	db := 0
	if v := env.MustEnv("REDIS_DB", false); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			db = parsed
		}
	}
	redisConfig := &RedisConfig{Addr: url, Password: password, DB: db}
	redisClient := NewRedisClient(*redisConfig)
	return redisClient
}
