package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Storage    StorageConfig
	Email      EmailConfig
	App        AppConfig
	Encryption EncryptionConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type StorageConfig struct {
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3Bucket           string
	MaxUploadSize      int64
	AllowedFileTypes   []string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

type AppConfig struct {
	FrontendURL string
}

type EncryptionConfig struct {
	Key string
}

// Load loads and validates the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "controlwise"),
			Password: getEnv("DB_PASSWORD", "controlwise"),
			DBName:   getEnv("DB_NAME", "controlwise"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"), // Required, no default
			Expiry: 24 * time.Hour,
		},
		Storage: StorageConfig{
			AWSRegion:          getEnv("AWS_REGION", "eu-west-1"),
			AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			S3Bucket:           getEnv("S3_BUCKET", "controlwise-files"),
			MaxUploadSize:      getEnvAsInt64("MAX_UPLOAD_SIZE", 10485760), // 10MB default
			AllowedFileTypes:   getEnvAsStringSlice("ALLOWED_FILE_TYPES", "image/jpeg,image/png,image/webp,application/pdf"),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			SMTPFrom:     getEnv("SMTP_FROM", "noreply@controlwise.io"),
		},
		App: AppConfig{
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", ""), // Required for storing Twilio credentials
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration is present and valid
func (c *Config) Validate() error {
	// JWT secret is required and must be at least 32 characters
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}
	if len(c.JWT.Secret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters for security")
	}

	// Validate environment
	validEnvs := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvs[c.Server.Env] {
		return fmt.Errorf("invalid ENV value: %s (must be development, staging, or production)", c.Server.Env)
	}

	// In production, SSL should be enabled for database
	if c.Server.Env == "production" && c.Database.SSLMode == "disable" {
		return errors.New("DB_SSL_MODE must not be 'disable' in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}
