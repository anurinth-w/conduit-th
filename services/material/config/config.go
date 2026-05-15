package config

import "os"

type Config struct {
	DatabaseURL string
	RedisURL    string
	Env         string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		Env:         getEnv("GO_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
