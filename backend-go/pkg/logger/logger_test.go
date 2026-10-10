package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskPII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Email masking",
			input:    "Contact john.doe@example.com for help",
			expected: "Contact j***@example.com for help",
		},
		{
			name:     "Phone masking",
			input:    "Call 123-456-7890 immediately",
			expected: "Call ***-***-**** immediately",
		},
		{
			name:     "Bearer token masking",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "Authorization: Bearer [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := MaskPII(tt.input)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestLoggerInitializationAndLogging(t *testing.T) {
	l := InitGlobalLogger("debug")
	assert.NotNil(t, l)

	assert.NotPanics(t, func() {
		Debug("test debug message", "key", "val")
		Info("test info message", "user_email", "admin@neighbor.com")
		Warn("test warn message", "retry_count", 2)
		Error("test error message", "error_code", 500)
		LogCtx(context.Background(), slog.LevelInfo, "test log ctx")
		_ = With("app", "neighbor-test")
	})
}
