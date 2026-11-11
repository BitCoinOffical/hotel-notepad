package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger(level string) (*zap.Logger, error) {
	switch level {
	case "dev":
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("zap.NewDevelopment: %w", err)
		}
		return logger, nil
	case "prod":
		logger, err := zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("zap.NewProduction: %w", err)
		}
		return logger, nil
	default:
		return nil, fmt.Errorf("invalid debug level")
	}
}
