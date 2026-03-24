package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenManager struct {
	redisClient *redis.Client
}

func NewTokenManager(redisClient *redis.Client) *TokenManager {
	return &TokenManager{redisClient: redisClient}
}

func (tm *TokenManager) SaveToken(userID int64, token string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("token:%d", userID)
	return tm.redisClient.Set(ctx, key, token, expiration).Err()
}

func (tm *TokenManager) ValidateToken(userID int64, token string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("token:%d", userID)

	storedToken, err := tm.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return storedToken == token, nil
}

func (tm *TokenManager) DeleteToken(userID int64) error {
	ctx := context.Background()
	key := fmt.Sprintf("token:%d", userID)
	return tm.redisClient.Del(ctx, key).Err()
}
