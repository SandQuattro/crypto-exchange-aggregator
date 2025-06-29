package logging

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	scopeFieldName   = "scope"
	traceIdFieldName = "trace_id"
)

var logger zerolog.Logger

func GetCtxLogger(ctx context.Context) zerolog.Logger {
	return logger.With().Ctx(ctx).Logger()
}

func InitLogger(debug bool) {
	partsOrder := []string{
		zerolog.LevelFieldName,
		zerolog.TimestampFieldName,
		traceIdFieldName,
		scopeFieldName,
		zerolog.MessageFieldName,
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		PartsOrder: partsOrder,
		FormatPrepare: func(m map[string]any) error {
			formatFieldValue[string](m, "%s", traceIdFieldName)
			formatFieldValue[string](m, "[%s]", scopeFieldName)
			return nil
		},
		FieldsExclude: []string{traceIdFieldName, scopeFieldName},
	}

	logger = zerolog.New(consoleWriter).Hook(ctxHook{})
	if debug {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}
	logger = logger.With().Timestamp().Logger()
}

func formatFieldValue[T any](vs map[string]any, format string, field string) {
	if v, ok := vs[field].(T); ok {
		vs[field] = fmt.Sprintf(format, v)
	} else {
		vs[field] = ""
	}
}

type ctxHook struct{}

func (h ctxHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	if scope, ok := GetScopeFromCtx(e.GetCtx()); ok {
		e.Str(scopeFieldName, scope)
	}
	if traceId, ok := GetTraceIdFromCtx(e.GetCtx()); ok {
		e.Str(traceIdFieldName, traceId)
	}
}

type scopeCtxKey struct{}

func GetCtxWithScope(ctx context.Context, scope string) context.Context {
	return context.WithValue(ctx, scopeCtxKey{}, scope)
}

type traceIdCtxKey struct{}

func GetCtxWithTraceId(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceIdCtxKey{}, generateTraceId())
}

func GetScopeFromCtx(ctx context.Context) (string, bool) {
	if scope, ok := ctx.Value(scopeCtxKey{}).(string); ok {
		return scope, true
	}
	return "", false
}

func GetTraceIdFromCtx(ctx context.Context) (string, bool) {
	if traceId, ok := ctx.Value(traceIdCtxKey{}).(string); ok {
		return traceId, true
	}
	return "", false
}

func generateTraceId() string {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return ""
	}

	return newUUID.String()
}
