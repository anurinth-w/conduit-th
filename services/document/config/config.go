package config

import "os"

type Config struct {
	DatabaseURL      string
	Env              string
	MinIOEndpoint    string
	MinIOAccessKey   string
	MinIOSecretKey   string
	MinIOBucket      string
	MinIOUseSSL      bool
	GotenbergURL     string
	PresignExpirySec int
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		Env:              getEnv("GO_ENV", "development"),
		MinIOEndpoint:    getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:   getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:   getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:      getEnv("MINIO_BUCKET", "conduit-documents"),
		MinIOUseSSL:      getEnv("MINIO_USE_SSL", "false") == "true",
		GotenbergURL:     getEnv("GOTENBERG_URL", "http://localhost:3000"),
		PresignExpirySec: 3600,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
