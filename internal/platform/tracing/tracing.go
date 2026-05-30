// internal/platform/tracing/tracing.go
package tracing

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Config struct {
	Enabled     bool
	Exporter    string
	Endpoint    string
	ServiceName string
}

type ShutdownFunc func(context.Context) error

func Init(ctx context.Context, cfg Config) (ShutdownFunc, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
		resource.WithProcess(),
		resource.WithOS(),
	)
	if err != nil {
		return nil, fmt.Errorf("tracing resource: %w", err)
	}

	var exporter sdktrace.SpanExporter
	if cfg.Exporter == "otlp" {
		endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "https://"), "http://")
		opts := []otlptracehttp.Option{otlptracehttp.WithInsecure()}
		if endpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
	} else {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
	if err != nil {
		return nil, fmt.Errorf("tracing exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
