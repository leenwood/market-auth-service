// internal/platform/logger/logger.go
package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey int

const (
	requestIDKey contextKey = iota
	traceIDKey
)

func New(level, format string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	l := base
	if rid := RequestIDFromContext(ctx); rid != "" {
		l = l.With("request_id", rid)
	}
	if tid := TraceIDFromContext(ctx); tid != "" {
		l = l.With("trace_id", tid)
	}
	return l
}
