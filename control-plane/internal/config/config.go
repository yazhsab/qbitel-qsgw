package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port        string `json:"port"`
	DatabaseURL string `json:"database_url"`
	EtcdAddr    string `json:"etcd_addr"`
	LogLevel    string `json:"log_level"`

	// Auth configuration
	JWTSecret    string   `json:"jwt_secret"`
	JWTIssuer    string   `json:"jwt_issuer"`
	APIKeys      []string `json:"api_keys"` // format: "key:subject:role"
	CORSOrigins  []string `json:"cors_origins"`
	MaxBodyBytes int64    `json:"max_body_bytes"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Port:         getEnv("QSGW_PORT", "8085"),
		DatabaseURL:  getEnv("QSGW_DATABASE_URL", ""),
		EtcdAddr:     getEnv("QSGW_ETCD_ADDR", "http://127.0.0.1:2379"),
		LogLevel:     getEnv("QSGW_LOG_LEVEL", "info"),
		JWTSecret:    getEnv("QUANTUN_JWT_SECRET", ""),
		JWTIssuer:    getEnv("QUANTUN_JWT_ISSUER", "quantun"),
		MaxBodyBytes: 1 << 20, // 1 MB
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("QSGW_DATABASE_URL is required")
	}

	// Parse API keys (comma-separated, format: key:subject:role)
	if apiKeysStr := getEnv("QUANTUN_API_KEYS", ""); apiKeysStr != "" {
		cfg.APIKeys = strings.Split(apiKeysStr, ",")
	}

	// Parse CORS origins (comma-separated)
	if corsStr := getEnv("QUANTUN_CORS_ORIGINS", ""); corsStr != "" {
		cfg.CORSOrigins = strings.Split(corsStr, ",")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
