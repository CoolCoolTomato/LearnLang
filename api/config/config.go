package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	User     UserConfig     `mapstructure:"user"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type UserConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "learnlang")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using defaults: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal("Failed to unmarshal config:", err)
	}

	return &cfg
}
