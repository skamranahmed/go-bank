package telemetry

import (
	"context"

	"github.com/skamranahmed/go-bank/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func InitTracer() (*trace.TracerProvider, error) {
	ctx := context.Background()

	telemetryConfig := config.GetTelemetryConfig()
	telemetryServiceName := telemetryConfig.ServiceName
	telemetryTracesIntakeEndpoint := telemetryConfig.TracesIntakeEndpoint

	// OTLP http exporter is responsible for sending data to the Elastic APM server
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(telemetryTracesIntakeEndpoint),
		otlptracehttp.WithURLPath("/v1/traces"), // trace intake endpoint for APM: https://www.elastic.co/docs/solutions/observability/apm/opentelemetry-intake-api
		otlptracehttp.WithInsecure(),            // TODO: for local this should be fine, need to use TLS in production
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(telemetryServiceName),
			semconv.DeploymentEnvironment(config.GetEnvironment()),
		),
	)
	if err != nil {
		return nil, err
	}

	// create tracer provider
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()), // sample all traces
	)

	// set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}
