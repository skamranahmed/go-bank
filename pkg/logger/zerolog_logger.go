package logger

import (
	"context"
	"fmt"
	"io"
	"log"
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

	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000"
	zerolog.TimestampFunc = func() time.Time {
		// TODO: I feel this should be better kept in UTC instead of local time (i.e IST). I will think about this.
		return time.Now().UTC().Add(5*time.Hour + 30*time.Minute)
	}

	output = os.Stderr

	if config.GetEnvironment() == config.APP_ENVIRONMENT_LOCAL {
		// ConsoleWriter is intended for use in local/development environments only.
		// In production, it can significantly slow down response times because it
		// parses JSON logs into plain text for human-readable output.
		// This becomes a potential bottleneck under high traffic conditions.
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: zerolog.TimeFieldFormat,
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

		/*
			If the server is running in the local environment, we want to write logs
			both to the console (for the developer to see in real-time) and to a file (for filebeat
			ingestion). The reason for writing logs to a file even in the local environment is as follows:

			On non-Linux host OSes (e.g. macOS or Windows), docker containers run inside a lightweight
			VM. This VM layer prevents the container's stdout/stderr logs from being easily accessed
			by Filebeat. As a result, if we rely solely on console logging inside the container,
			filebeat may not be able to stream these logs to logstash

			To solve this, we create a dedicated log directory and write logs to a file inside it.
			This file can then be safely picked up by filebeat, ensuring that log shipping works
			consistently regardless of the host OS or whether the server is running natively or in docker

			The serverLogsDirectory is chosen based on whether the server is running in a container:
			  - If inside docker (detected via /.dockerenv), logs are written to /logs (mounted volume)
			  - If running natively on the host OS, logs are written to ./native-logs

			Reference: https://www.exoscale.com/syslog/docker-logging/
		*/
		serverLogsDirectory := "./native-logs"
		// check whether the server is running natively on the host os or in a dockerized container
		_, err := os.Stat("/.dockerenv")
		if err == nil {
			// this means the server is running in a dockerized container
			serverLogsDirectory = "/logs"
		}

		err = os.MkdirAll(serverLogsDirectory, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}

		logFile, err := os.OpenFile(filepath.Join(serverLogsDirectory, "api.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}

		// MultiWriter writes to both console/stderr and a file
		output = io.MultiWriter(output, logFile)
	}

	logLevel := zerolog.InfoLevel
	if config.GetLoggerConfig().Level == config.LogLevelDebug {
		logLevel = zerolog.DebugLevel
	}

	zlogger := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Str("role", config.Role). // this will be helpful to filter out logs from server or worker in Kibana
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
