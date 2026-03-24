package database

import (
	"context"
	"fmt"
	"learnlang-api/config"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis(cfg *config.Config) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("Redis connected successfully")
	return nil
}
