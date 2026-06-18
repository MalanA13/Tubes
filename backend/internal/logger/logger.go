package logger

import (
	"os"

	"go.uber.org/zap"
)

// InitLogger initializes a global zap logger based on the APP_ENV.
// It also attaches the service_name field to all structured logs.
func InitLogger(serviceName string) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	env := os.Getenv("APP_ENV")
	if env == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		return nil, err
	}

	// Attach service_name
	logger = logger.With(zap.String("service_name", serviceName))

	// Replace standard global logger
	zap.ReplaceGlobals(logger)

	return logger, nil
}
