package observability

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type LogEntry struct {
	Timestamp string         `json:"timestamp"`
	Level     Level          `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

type Logger struct {
	minLevel Level
}

var defaultLogger = NewLogger()

func NewLogger() *Logger {
	return &Logger{minLevel: LevelInfo}
}

func DefaultLogger() *Logger {
	return defaultLogger
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		log.Printf("logger marshal error: %v", err)
		return
	}
	_, _ = os.Stdout.Write(append(b, '\n'))
}

func (l *Logger) Info(msg string, fields ...map[string]any) {
	l.log(LevelInfo, msg, mergeFields(fields...))
}

func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...map[string]any) {
	l.log(LevelInfo, msg, withContextFields(ctx, mergeFields(fields...)))
}

func (l *Logger) Warn(msg string, fields ...map[string]any) {
	l.log(LevelWarn, msg, mergeFields(fields...))
}

func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...map[string]any) {
	l.log(LevelWarn, msg, withContextFields(ctx, mergeFields(fields...)))
}

func (l *Logger) Error(msg string, fields ...map[string]any) {
	l.log(LevelError, msg, mergeFields(fields...))
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...map[string]any) {
	l.log(LevelError, msg, withContextFields(ctx, mergeFields(fields...)))
}

func Info(ctx context.Context, msg string, fields ...map[string]any) {
	DefaultLogger().InfoContext(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...map[string]any) {
	DefaultLogger().WarnContext(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...map[string]any) {
	DefaultLogger().ErrorContext(ctx, msg, fields...)
}

func mergeFields(fields ...map[string]any) map[string]any {
	merged := make(map[string]any)
	for _, f := range fields {
		for key, value := range f {
			merged[key] = value
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func withContextFields(ctx context.Context, fields map[string]any) map[string]any {
	correlationID, ok := CorrelationIDFromContext(ctx)
	if !ok || correlationID == "" {
		return fields
	}

	out := cloneFields(fields)
	if _, exists := out["correlation_id"]; !exists {
		out["correlation_id"] = correlationID
	}
	if _, exists := out["request_id"]; !exists {
		out["request_id"] = correlationID
	}
	return out
}

func cloneFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return make(map[string]any)
	}

	out := make(map[string]any, len(fields))
	for key, value := range fields {
		out[key] = value
	}
	return out
}
