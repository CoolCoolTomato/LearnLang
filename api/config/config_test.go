package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadDefaults(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDirectory) })

	cfg := Load()
	if cfg.Database.Host != "localhost" || cfg.Database.Port != "5432" || cfg.Database.DBName != "learnlang" {
		t.Fatalf("database defaults = %#v", cfg.Database)
	}
	if cfg.Server.Port != "8080" || cfg.JWT.Secret != "your-secret-key" {
		t.Fatalf("server/JWT defaults = %#v / %#v", cfg.Server, cfg.JWT)
	}
	if cfg.Redis.Host != "localhost" || cfg.Redis.Port != "6379" || cfg.Redis.DB != 0 {
		t.Fatalf("redis defaults = %#v", cfg.Redis)
	}
	if cfg.Milvus.Collection != "user_memory_summaries" {
		t.Fatalf("milvus defaults = %#v", cfg.Milvus)
	}
}

func TestLoadConfigFile(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	directory := t.TempDir()
	configYAML := []byte("server:\n  port: '9090'\ndatabase:\n  host: db.internal\nredis:\n  db: 4\n")
	if err := os.WriteFile(directory+string(os.PathSeparator)+"config.yaml", configYAML, 0o600); err != nil {
		t.Fatal(err)
	}
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDirectory) })

	cfg := Load()
	if cfg.Server.Port != "9090" || cfg.Database.Host != "db.internal" || cfg.Redis.DB != 4 {
		t.Fatalf("Load() = %#v", cfg)
	}
	if cfg.Database.Port != "5432" {
		t.Fatalf("default was not retained: database port = %q", cfg.Database.Port)
	}
}
