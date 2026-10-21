package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseUrl string
	Port        string
	JWTSecret   string

	// Database pool settings
	DBMaxConns          int32
	DBMinConns          int32
	DBMaxConnLifetime   time.Duration
	DBMaxConnIdleTime   time.Duration
	DBHealthCheckPeriod time.Duration

	// Rate limiting
	RateLimitRPS   int
	RateLimitBurst int

	// Server timeouts
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// Request limits
	MaxRequestBodySize int64

	// CORS
	AllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseUrl: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	if cfg.DatabaseUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	if cfg.Port == "" {
		cfg.Port = "8000"
	}

	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "super-secret-dev-key-change-in-production"
	}

	// Database pool defaults – tuned for 50k users
	cfg.DBMaxConns = int32(getEnvInt("DB_MAX_CONNS", 100))
	cfg.DBMinConns = int32(getEnvInt("DB_MIN_CONNS", 10))
	cfg.DBMaxConnLifetime = getEnvDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute)
	cfg.DBMaxConnIdleTime = getEnvDuration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute)
	cfg.DBHealthCheckPeriod = getEnvDuration("DB_HEALTH_CHECK_PERIOD", 30*time.Second)

	// Rate limiting defaults
	cfg.RateLimitRPS = getEnvInt("RATE_LIMIT_RPS", 100)
	cfg.RateLimitBurst = getEnvInt("RATE_LIMIT_BURST", 200)

	// Server timeouts
	cfg.ReadTimeout = getEnvDuration("READ_TIMEOUT", 15*time.Second)
	cfg.WriteTimeout = getEnvDuration("WRITE_TIMEOUT", 15*time.Second)
	cfg.IdleTimeout = getEnvDuration("IDLE_TIMEOUT", 120*time.Second)
	cfg.ShutdownTimeout = getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second)

	// Request body limit (default 1MB)
	cfg.MaxRequestBodySize = int64(getEnvInt("MAX_REQUEST_BODY_SIZE", 1_048_576))

	// CORS origins
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		cfg.AllowedOrigins = splitAndTrim(origins)
	} else {
		cfg.AllowedOrigins = []string{"*"}
	}

	return cfg, nil
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range splitString(s, ",") {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	var parts []string
	for {
		i := indexOf(s, sep)
		if i < 0 {
			parts = append(parts, s)
			break
		}
		parts = append(parts, s[:i])
		s = s[i+len(sep):]
	}
	return parts
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
