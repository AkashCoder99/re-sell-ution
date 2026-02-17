package observability

import (
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

func NewLogger() *Logger {
	return &Logger{minLevel: LevelInfo}
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

func (l *Logger) Warn(msg string, fields ...map[string]any) {
	l.log(LevelWarn, msg, mergeFields(fields...))
}

func (l *Logger) Error(msg string, fields ...map[string]any) {
	l.log(LevelError, msg, mergeFields(fields...))
}

func mergeFields(fields ...map[string]any) map[string]any {
	for _, f := range fields {
		if len(f) > 0 {
			return f
		}
	}
	return nil
}
