// Package postgres provides PostgreSQL database connectivity for production services.
// It reads connection parameters exclusively from environment variables.
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
	"os"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open creates a new PostgreSQL database connection using environment variables.
// This function replaces the SQLite Open() used in production cmd entrypoints.
func Open() (*gorm.DB, error) {
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