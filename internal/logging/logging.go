package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type ctxKey struct{}

// Into returns a child context carrying l.
func Into(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From returns the logger carried by ctx, or a no-op logger if none was set.
func From(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	}
	return zap.NewNop()
}

// WithTrace returns l enriched with trace_id/span_id
func WithTrace(ctx context.Context, l *zap.Logger) *zap.Logger {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		return l.With(
			zap.String("trace_id", sc.TraceID().String()),
			zap.String("span_id", sc.SpanID().String()),
		)
	}
	return l
}
