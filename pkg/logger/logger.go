package logger

import "sync"

var (
	once           sync.Once
	loggerInstance Logger
)

type Logger interface {
	Info(message string, args ...any)
	InfoFields(message string, fields map[string]any)
	Warn(message string, args ...any)
	Error(message string, args ...any)
	Fatal(format string, args ...any)
}

func Init() {
	once.Do(func() {
		loggerInstance = newZerologLogger()
	})
}

func Info(message string, args ...any) {
	loggerInstance.Info(message, args...)
}

func InfoFields(message string, fields map[string]any) {
	loggerInstance.InfoFields(message, fields)
}

func Warn(message string, args ...any) {
	loggerInstance.Warn(message, args...)
}

func Error(message string, args ...any) {
	loggerInstance.Error(message, args...)
}

func Fatal(message string, args ...any) {
	loggerInstance.Fatal(message, args...)
}
