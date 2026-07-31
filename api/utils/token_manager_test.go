package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTokenManagerWithMockRedis(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	manager := NewTokenManager(client)

	if err := manager.SaveToken(7, "token", time.Hour); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	if valid, err := manager.ValidateToken(7, "token"); err != nil || !valid {
		t.Fatalf("ValidateToken(correct) = %v, %v", valid, err)
	}
	if valid, err := manager.ValidateToken(7, "other"); err != nil || valid {
		t.Fatalf("ValidateToken(wrong) = %v, %v", valid, err)
	}
	if valid, err := manager.ValidateToken(8, "token"); err != nil || valid {
		t.Fatalf("ValidateToken(missing) = %v, %v", valid, err)
	}
	if err := manager.DeleteToken(7); err != nil {
		t.Fatalf("DeleteToken() error = %v", err)
	}
	if server.Exists("token:7") {
		t.Fatal("DeleteToken() left the Redis key")
	}
}

func TestTokenManagerPropagatesRedisErrors(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	manager := NewTokenManager(client)
	server.Close()

	if err := manager.SaveToken(1, "token", time.Hour); err == nil {
		t.Fatal("SaveToken() did not return a connection error")
	}
	if valid, err := manager.ValidateToken(1, "token"); err == nil || valid {
		t.Fatalf("ValidateToken() = %v, %v", valid, err)
	}
	if err := manager.DeleteToken(1); err == nil {
		t.Fatal("DeleteToken() did not return a connection error")
	}
	_ = client.Close()
	if !errors.Is(client.Ping(t.Context()).Err(), redis.ErrClosed) {
		t.Fatal("test Redis client did not close")
	}
}
