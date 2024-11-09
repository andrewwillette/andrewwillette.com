package server

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create an OTLP exporter for sending traces
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	// Set up the tracer provider with the exporter and default options
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("audio-app"),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
	)

	// Register the global tracer provider
	otel.SetTracerProvider(tp)
	return tp, nil
}
