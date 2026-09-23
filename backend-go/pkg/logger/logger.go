package logger

import (
	"context"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`(?i)([a-zA-Z0-9_.+-])[a-zA-Z0-9_.+-]+@([a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)`)
	phoneRegex = regexp.MustCompile(`\b(?:\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`)
	tokenRegex = regexp.MustCompile(`(?i)(bearer\s+|token\s+|key=)([a-zA-Z0-9_\-\.]{10,})`)
)

// MaskPII sanitizes sensitive personal information from strings before writing to logs.
func MaskPII(input string) string {
	if input == "" {
		return ""
	}
	// Mask email addresses (e.g. j***@example.com)
	out := emailRegex.ReplaceAllStringFunc(input, func(m string) string {
		parts := strings.Split(m, "@")
		if len(parts) != 2 || len(parts[0]) == 0 {
			return m
		}
		return string(parts[0][0]) + "***@" + parts[1]
	})

	// Mask phone numbers
	out = phoneRegex.ReplaceAllString(out, "***-***-****")

	// Mask auth tokens / API keys
	out = tokenRegex.ReplaceAllString(out, "$1[REDACTED]")

	return out
}

// InitGlobalLogger initializes Go's standard slog logger with structured JSON output and PII masking.
func InitGlobalLogger(env string) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(env) {
	case "production", "prod":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Mask string values for PII
			if a.Value.Kind() == slog.KindString {
				return slog.String(a.Key, MaskPII(a.Value.String()))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

// Info logs an informational message.
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// With creates a child logger with key-value pairs.
func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

// LogCtx logs with context.
func LogCtx(ctx context.Context, level slog.Level, msg string, args ...any) {
	slog.Log(ctx, level, msg, args...)
}
