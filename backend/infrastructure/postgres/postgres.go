// Package postgres provides PostgreSQL database connectivity for production services.
// It reads connection parameters from process environment variables and optional local
// .env files for development. Existing process environment variables are never
// overwritten by .env values, preserving Docker and production configuration precedence.
//
// Required environment variables:
//   - DB_HOST     (default: localhost)
//   - DB_PORT     (default: 5432)
//   - DB_NAME     (default: logistics)
//   - DB_USER     (default: logistics)
//   - DB_PASSWORD (default: logistics)
//   - DB_SSLMODE  (default: disable)
package postgres

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open creates a new PostgreSQL database connection using environment variables.
// This function replaces the SQLite Open() used in production cmd entrypoints.
func Open() (*gorm.DB, error) {
	LoadEnv()
	log.Printf("PostgreSQL configuration loaded: DB_HOST=%s DB_PORT=%s DB_NAME=%s DB_USER=%s DB_SSLMODE=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "logistics"),
		getEnv("DB_USER", "logistics"),
		getEnv("DB_SSLMODE", "disable"),
	)

	dsn := BuildDSN()
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Verify connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return db, nil
}

// LoadEnv loads local development environment files when they exist.
//
// LoadEnv intentionally uses godotenv.Load rather than godotenv.Overload so that
// environment variables provided by Docker, CI, or production hosts keep precedence
// over values from backend/.env. The first existing path among ".env" and
// "backend/.env" is loaded to support running commands from either the backend
// directory or the repository root.
func LoadEnv() {
	for _, path := range []string{".env", "backend/.env"} {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("failed to load %s: %v", path, err)
			}
			return
		}
	}
}

// BuildDSN constructs a PostgreSQL DSN string from environment variables.
func BuildDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "logistics"),
		getEnv("DB_USER", "logistics"),
		getEnv("DB_PASSWORD", "logistics"),
		getEnv("DB_SSLMODE", "disable"),
	)
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}