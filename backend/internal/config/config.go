package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Log      LogConfig
	WebSocket WebSocketConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Mode         string // debug, release, test
}

type DatabaseConfig struct {
	URL            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxLifetime time.Duration
}

type AuthConfig struct {
	Secret           string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	PingPeriod      time.Duration
	PongWait        time.Duration
}

// Load loads configuration from environment variables.
// JWT_SECRET is REQUIRED and never given a default: running without it would
// let anyone forge tokens, so we fail fast instead of silently shipping unsafe.
func Load() (*Config, error) {
	port, _ := strconv.Atoi(getEnv("SERVER_PORT", "8081"))
	logLevel := getEnv("LOG_LEVEL", "info")
	dbURL := getEnv("DATABASE_URL", "sqlite://data/aicenter.db")
	authSecret := os.Getenv("JWT_SECRET")
	if authSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required (do not hard-code a default)")
	}

	return &Config{
		Server: ServerConfig{
			Port:         port,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			Mode:         getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			URL:             dbURL,
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Auth: AuthConfig{
			Secret:          authSecret,
			AccessTokenTTL:  60 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Log: LogConfig{
			Level:  logLevel,
			Format: getEnv("LOG_FORMAT", "json"),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			PingPeriod:      30 * time.Second,
			PongWait:        60 * time.Second,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
