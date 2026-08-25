package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	AdminPort  string
	ServerEnv  string

	DBDriver     string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSQLitePath string

	JWTSecret     string
	JWTExpiration int

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string

	SessionSecret string
	SessionMaxAge int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRATION", "86400"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}

	sessionMaxAge, err := strconv.Atoi(getEnv("SESSION_MAX_AGE", "604800"))
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_MAX_AGE: %w", err)
	}

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		AdminPort:  getEnv("ADMIN_PORT", "9080"),
		ServerEnv:  getEnv("SERVER_ENV", "development"),

		DBDriver:     getEnv("DB_DRIVER", "mysql"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "root"),
		DBPassword:   getEnv("DB_PASSWORD", ""),
		DBName:       getEnv("DB_NAME", "saas_oauth2"),
		DBSQLitePath: getEnv("DB_SQLITE_PATH", "./useria.db"),

		JWTSecret:     mustGetEnv("JWT_SECRET"),
		JWTExpiration: jwtExp,

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),

		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURL:  getEnv("GITHUB_REDIRECT_URL", ""),

		SessionSecret: mustGetEnv("SESSION_SECRET"),
		SessionMaxAge: sessionMaxAge,
	}, nil
}

func (c *Config) DSN() string {
	if c.DBDriver == "sqlite" {
		return c.DBSQLitePath
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) MigrationPath() string {
	if c.DBDriver == "sqlite" {
		return "internal/database/migrations/001_init_sqlite.sql"
	}
	return "internal/database/migrations/001_init.sql"
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}
