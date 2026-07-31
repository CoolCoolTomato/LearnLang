package database

import (
	"learnlang-api/config"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestMilvusAddress(t *testing.T) {
	for _, tt := range []struct {
		cfg  config.MilvusConfig
		want string
	}{
		{config.MilvusConfig{}, "localhost:19530"},
		{config.MilvusConfig{Host: "milvus", Port: "1234"}, "milvus:1234"},
		{config.MilvusConfig{Host: "milvus:9999", Port: "1234"}, "milvus:9999"},
	} {
		if got := milvusAddress(tt.cfg); got != tt.want {
			t.Errorf("milvusAddress(%#v) = %q, want %q", tt.cfg, got, tt.want)
		}
	}
}

func TestConnectRedisWithMockServer(t *testing.T) {
	server := miniredis.RunT(t)
	host, port, ok := strings.Cut(server.Addr(), ":")
	if !ok {
		t.Fatalf("invalid mock Redis address %q", server.Addr())
	}
	previous := RedisClient
	t.Cleanup(func() {
		if RedisClient != nil {
			_ = RedisClient.Close()
		}
		RedisClient = previous
	})
	cfg := &config.Config{Redis: config.RedisConfig{Host: host, Port: port}}
	if err := ConnectRedis(cfg); err != nil {
		t.Fatalf("ConnectRedis() error = %v", err)
	}
	if err := RedisClient.Set(t.Context(), "key", "value", 0).Err(); err != nil {
		t.Fatalf("connected Redis client did not write: %v", err)
	}
	got, getErr := server.Get("key")
	if getErr != nil || got != "value" {
		t.Fatalf("mock Redis value = %q, %v", got, getErr)
	}
}
