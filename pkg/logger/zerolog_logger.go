package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/skamranahmed/go-bank/config"
)

type zerologLogger struct {
	logger zerolog.Logger
}

func newZerologLogger() Logger {
	var output io.Writer

	// ConsoleWriter is intended for use in local/development environments only.
	// In production, it can significantly slow down response times because it
	// parses JSON logs into plain text for human-readable output.
	// This becomes a potential bottleneck under high traffic conditions.
	output = zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i any) string {
			return strings.ToUpper(fmt.Sprintf("[%s] |", i))
		},
		FormatMessage: func(i any) string {
			return fmt.Sprintf("| %s", i)
		},
		FormatCaller: func(i any) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		PartsExclude: []string{
			// We can exclude fields from the logs if we want
			// zerolog.TimestampFieldName,
		},
	}

	if config.GetEnvironment() != config.APP_ENVIRONMENT_LOCAL {
		output = os.Stderr
	}

	logLevel := zerolog.InfoLevel
	if config.GetLoggerConfig().Level == config.LogLevelDebug {
		logLevel = zerolog.DebugLevel
	}

	zlogger := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		CallerWithSkipFrameCount(4).
		Logger()

	return &zerologLogger{
		logger: zlogger,
	}
}

func (z *zerologLogger) Info(ctx context.Context, message string, args ...any) {
	correlationID := z.extractCorrelationIDFromCtx(ctx)
	z.logger.Info().Any("correlation_id", correlationID).Msgf(message, args...)
}

func (z *zerologLogger) InfoFields(message string, fields map[string]any) {
	z.logger.Info().Fields(fields).Msg(message)
}

func (z *zerologLogger) Warn(ctx context.Context, message string, args ...any) {
	correlationID := z.extractCorrelationIDFromCtx(ctx)
	z.logger.Warn().Any("correlation_id", correlationID).Msgf(message, args...)
}

func (z *zerologLogger) Error(ctx context.Context, message string, args ...any) {
	correlationID := z.extractCorrelationIDFromCtx(ctx)
	z.logger.Error().Any("correlation_id", correlationID).Msgf(message, args...)
}

func (z *zerologLogger) Fatal(ctx context.Context, message string, args ...any) {
	correlationID := z.extractCorrelationIDFromCtx(ctx)
	z.logger.Fatal().Any("correlation_id", correlationID).Msgf(message, args...)
}

func (z *zerologLogger) extractCorrelationIDFromCtx(ctx context.Context) any {
	return ctx.Value("correlation_id")
}
