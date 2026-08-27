package logger

import (
	"context"
	"encoding/json"
	"os"
	"time"
)

// LogLevel определяет уровень логирования
type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogEntry структура для JSON логов
type LogEntry struct {
	Time    string  `json:"time"`
	Level   string  `json:"level"`
	TraceID string  `json:"trace_id,omitempty"`
	Message string  `json:"message"`
	URL     string  `json:"url,omitempty"`
	Status  int     `json:"status,omitempty"`
	Error   string  `json:"error,omitempty"`
	DurMS   int64   `json:"duration_ms,omitempty"`
}

type contextKey string

const TraceIDKey contextKey = "trace_id"

// log пишет сообщение в stdout
func log(level LogLevel, traceID, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Time:    time.Now().UTC().Format(time.RFC3339Nano),
		Level:   string(level),
		TraceID: traceID,
		Message: message,
	}

	// Добавляем дополнительные поля
	for k, v := range fields {
		switch k {
		case "url":
			entry.URL = v.(string)
		case "status":
			entry.Status = v.(int)
		case "error":
			entry.Error = v.(string)
		case "duration_ms":
			entry.DurMS = v.(int64)
		}
	}

	json.NewEncoder(os.Stdout).Encode(entry)
}

// Публичные методы
func Info(ctx context.Context, message string, fields map[string]interface{}) {
	traceID := getTraceID(ctx)
	log(LevelInfo, traceID, message, fields)
}

func Warn(ctx context.Context, message string, fields map[string]interface{}) {
	traceID := getTraceID(ctx)
	log(LevelWarn, traceID, message, fields)
}

func Error(ctx context.Context, message string, fields map[string]interface{}) {
	traceID := getTraceID(ctx)
	log(LevelError, traceID, message, fields)
}

func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// NewContextWithTraceID создает контекст с trace_id
func NewContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID извлекает trace_id из контекста
func GetTraceID(ctx context.Context) string {
	return getTraceID(ctx)
}