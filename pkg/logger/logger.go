package logger

import "sync"

var (
	once           sync.Once
	loggerInstance Logger
)

type Logger interface {
	Infof(message string, args ...any)
	Warnf(message string, args ...any)
	Errorf(message string, args ...any)
	Fatalf(format string, args ...any)
}

func Init() {
	once.Do(func() {
		loggerInstance = newZerologLogger()
	})
}

func Infof(message string, args ...any) {
	loggerInstance.Infof(message, args...)
}

func Warnf(message string, args ...any) {
	loggerInstance.Warnf(message, args...)
}

func Errorf(message string, args ...any) {
	loggerInstance.Errorf(message, args...)
}

func Fatalf(message string, args ...any) {
	loggerInstance.Fatalf(message, args...)
}
