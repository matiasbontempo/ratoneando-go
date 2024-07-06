package cache

import (
	"context"
	"ratoneando/config"
	"ratoneando/utils/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
)

func Init() {
	logger.Log(config.REDIS_URL)
	opts, err := redis.ParseURL(config.REDIS_URL)
	if err != nil {
		logger.LogFatal("Error parsing Redis URL")
	}

	client := redis.NewClient(opts)

	Client = client
}

func Set(key string, value string, expiration int) error {
	ctx := context.Background()
	err := Client.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
	if err != nil {
		logger.LogWarn("Error setting key: " + key)
		return err
	}
	return nil
}

func Get(key string) (string, error) {
	ctx := context.Background()
	value, err := Client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}
