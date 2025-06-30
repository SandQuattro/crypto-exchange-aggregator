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
	traceIDFieldName = "trace_id"
)

var globalLogger zerolog.Logger //nolint:gochecknoglobals // Global logger is acceptable pattern

func GetCtxLogger(ctx context.Context) zerolog.Logger {
	return globalLogger.With().Ctx(ctx).Logger()
}

func InitLogger(debug bool) {
	partsOrder := []string{
		zerolog.LevelFieldName,
		zerolog.TimestampFieldName,
		traceIDFieldName,
		scopeFieldName,
		zerolog.MessageFieldName,
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		PartsOrder: partsOrder,
		FormatPrepare: func(m map[string]any) error {
			formatFieldValue[string](m, "%s", traceIDFieldName)
			formatFieldValue[string](m, "[%s]", scopeFieldName)
			return nil
		},
		FieldsExclude: []string{traceIDFieldName, scopeFieldName},
	}

	globalLogger = zerolog.New(consoleWriter).Hook(ctxHook{})
	if debug {
		globalLogger = globalLogger.Level(zerolog.DebugLevel)
	} else {
		globalLogger = globalLogger.Level(zerolog.InfoLevel)
	}
	globalLogger = globalLogger.With().Timestamp().Logger()
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
	if traceID, ok := GetTraceIDFromCtx(e.GetCtx()); ok {
		e.Str(traceIDFieldName, traceID)
	}
}

type scopeCtxKey struct{}

func GetCtxWithScope(ctx context.Context, scope string) context.Context {
	return context.WithValue(ctx, scopeCtxKey{}, scope)
}

type traceIDCtxKey struct{}

func GetCtxWithTraceID(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceIDCtxKey{}, generateTraceID())
}

func GetScopeFromCtx(ctx context.Context) (string, bool) {
	if scope, ok := ctx.Value(scopeCtxKey{}).(string); ok {
		return scope, true
	}
	return "", false
}

func GetTraceIDFromCtx(ctx context.Context) (string, bool) {
	if traceID, ok := ctx.Value(traceIDCtxKey{}).(string); ok {
		return traceID, true
	}
	return "", false
}

func generateTraceID() string {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return ""
	}

	return newUUID.String()
}
