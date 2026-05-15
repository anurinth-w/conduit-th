package config

import "os"

type Config struct {
	Env       string
	JWTSecret string
	Port      string

	// upstream services
	AuthURL     string
	UserURL     string
	JobURL      string
	MaterialURL string
	MediaURL    string
	DocumentURL string
	NotifyURL   string

	// rate limit
	RateLimitRPS   int
	RateLimitBurst int
}

func Load() *Config {
	return &Config{
		Env:       getEnv("GO_ENV", "development"),
		JWTSecret: getEnv("JWT_SECRET", "change-me"),
		Port:      getEnv("PORT", "8000"),

		AuthURL:     getEnv("AUTH_URL", "http://localhost:8001"),
		UserURL:     getEnv("USER_URL", "http://localhost:8002"),
		JobURL:      getEnv("JOB_URL", "http://localhost:8003"),
		MaterialURL: getEnv("MATERIAL_URL", "http://localhost:8004"),
		MediaURL:    getEnv("MEDIA_URL", "http://localhost:8005"),
		DocumentURL: getEnv("DOCUMENT_URL", "http://localhost:8006"),
		NotifyURL:   getEnv("NOTIFY_URL", "http://localhost:8007"),

		RateLimitRPS:   10,
		RateLimitBurst: 30,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
