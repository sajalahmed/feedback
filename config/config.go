package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                 string
	ServerPort             string
	DatabaseDSN            string
	JWTSecret              string
	JWTTokenExpireMinutes  int
	LoginLinkExpireMinutes int
	RateLimitSeconds       int
	AppURL                 string
	DeepLinkURL            string
	SMTP                   SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "root")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "feedback_app")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true", dbUser, dbPass, dbHost, dbPort, dbName)
	if val := os.Getenv("DATABASE_DSN"); val != "" {
		dsn = val
	}

	cfg := &Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		ServerPort:             getEnv("SERVER_PORT", ":8080"),
		DatabaseDSN:            dsn,
		JWTSecret:              getEnv("JWT_SECRET", "super-secret-key"),
		JWTTokenExpireMinutes:  getEnvInt("JWT_TOKEN_EXPIRE_MINUTES", 120),
		LoginLinkExpireMinutes: getEnvInt("LOGIN_LINK_EXPIRE_MINUTES", 15),
		RateLimitSeconds:       getEnvInt("RATE_LIMIT", 5),
		AppURL:                 getEnv("APP_URL", "http://localhost:8080"),
		DeepLinkURL:            getEnv("DEEPLINK_URL", "exp://127.0.0.1:8081/--/auth/callback"),
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getEnv("SMTP_PORT", "2525"),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@feedback.app"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET must be set")
	}
	if c.AppEnv == "production" && c.JWTSecret == "super-secret-key" {
		return fmt.Errorf("JWT_SECRET must be set to a non-default value in production")
	}
	if c.RateLimitSeconds <= 0 {
		return fmt.Errorf("RATE_LIMIT must be greater than zero")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
