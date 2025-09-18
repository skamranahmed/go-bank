package logger

import (
	"context"
	"sync"
)

var (
	once           sync.Once
	loggerInstance Logger
)

type Logger interface {
	Info(ctx context.Context, message string, args ...any)
	InfoFields(message string, fields map[string]any)
	Warn(ctx context.Context, message string, args ...any)
	Error(ctx context.Context, message string, args ...any)
	Fatal(ctx context.Context, message string, args ...any)
}

func Init() {
	once.Do(func() {
		loggerInstance = newZerologLogger()
	})
}

func Info(ctx context.Context, message string, args ...any) {
	loggerInstance.Info(ctx, message, args...)
}

func InfoFields(message string, fields map[string]any) {
	loggerInstance.InfoFields(message, fields)
}

func Warn(ctx context.Context, message string, args ...any) {
	loggerInstance.Warn(ctx, message, args...)
}

func Error(ctx context.Context, message string, args ...any) {
	loggerInstance.Error(ctx, message, args...)
}

func Fatal(ctx context.Context, message string, args ...any) {
	loggerInstance.Fatal(ctx, message, args...)
}
