package config

import (
"os"
"time"
)

type Config struct {
DatabaseURL    string
RedisURL       string
RedisPassword  string
JWTSecret      string
AccessTokenTTL time.Duration
RefreshTokenTTL time.Duration
Env            string
}

func Load() *Config {
accessTTL, _ := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
refreshTTL, _ := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))

return &Config{
DatabaseURL:     getEnv("DATABASE_URL", ""),
RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
RedisPassword:   getEnv("REDIS_PASSWORD", ""),
JWTSecret:       getEnv("JWT_SECRET", "change-me"),
AccessTokenTTL:  accessTTL,
RefreshTokenTTL: refreshTTL,
Env:             getEnv("GO_ENV", "development"),
}
}

func getEnv(key, fallback string) string {
if v := os.Getenv(key); v != "" {
return v
}
return fallback
}
