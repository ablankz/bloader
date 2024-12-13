// Package logger provides logging functionality for the application
package logger

import (
	"context"

	"github.com/ablankz/bloader/internal/config"
)

// Logger is an interface for logging
type Logger interface {
	SetupLogger(env string, conf config.ValidLoggingConfig) error
	With(args ...keyVal) Logger
	Debug(ctx context.Context, msg string, args ...keyVal)
	Info(ctx context.Context, msg string, args ...keyVal)
	Warn(ctx context.Context, msg string, args ...keyVal)
	Error(ctx context.Context, msg string, args ...keyVal)
	Close() error
}

// NewLoggerFromConfig creates a new Logger from the config
func NewLoggerFromConfig(env string, conf config.ValidLoggingConfig) (Logger, error) {
	return &SlogLogger{}, nil
}

type keyVal struct {
	Key   string
	Value any
}

// Value creates a new KeyVal
func Value(key string, value any) keyVal {
	return keyVal{Key: key, Value: value}
}

// Group creates a new KeyVal with a group of KeyVals
func Group(key string, kvs ...keyVal) keyVal {
	return keyVal{Key: key, Value: kvs}
}
