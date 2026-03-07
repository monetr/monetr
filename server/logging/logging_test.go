package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/monetr/monetr/server/internal/ctxkeys"
	"github.com/stretchr/testify/assert"
)

func TestParseLevel(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected slog.Level
		}{
			{"trace", "trace", LevelTrace},
			{"trace uppercase", "TRACE", LevelTrace},
			{"debug", "debug", slog.LevelDebug},
			{"debug uppercase", "DEBUG", slog.LevelDebug},
			{"info", "info", slog.LevelInfo},
			{"info uppercase", "INFO", slog.LevelInfo},
			{"warn", "warn", slog.LevelWarn},
			{"warning", "warning", slog.LevelWarn},
			{"error", "error", slog.LevelError},
			{"error uppercase", "ERROR", slog.LevelError},
			{"whitespace is trimmed", "  info  ", slog.LevelInfo},
			{"unknown input defaults to info", "unknown", slog.LevelInfo},
			{"empty input defaults to info", "", slog.LevelInfo},
		}

		for _, iteme := range tests {
			assert.Equal(t, iteme.expected, parseLevel(iteme.input))
		}
	})
}

func TestContextHandler(t *testing.T) {
	t.Run("injects context fields into log record", func(t *testing.T) {
		var buf bytes.Buffer
		inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: LevelTrace})
		logger := slog.New(NewContextHandler(inner))

		ctx := context.WithValue(context.Background(), ctxkeys.AccountID, uint64(1234))
		ctx = context.WithValue(ctx, ctxkeys.UserID, uint64(5678))

		logger.InfoContext(ctx, "test message")

		assert.True(t, json.Valid(buf.Bytes()), "output must be valid JSON")

		var record map[string]any
		assert.NoError(t, json.Unmarshal(buf.Bytes(), &record))
		assert.Equal(t, float64(1234), record["accountId"])
		assert.Equal(t, float64(5678), record["userId"])
	})

	t.Run("empty context produces no extra fields", func(t *testing.T) {
		var buf bytes.Buffer
		inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: LevelTrace})
		logger := slog.New(NewContextHandler(inner))

		logger.InfoContext(context.Background(), "test message")

		assert.True(t, json.Valid(buf.Bytes()), "output must be valid JSON")

		var record map[string]any
		assert.NoError(t, json.Unmarshal(buf.Bytes(), &record))
		assert.NotContains(t, record, "accountId")
		assert.NotContains(t, record, "userId")
		assert.NotContains(t, record, "requestId")
		assert.NotContains(t, record, "loginId")
		assert.NotContains(t, record, "jobId")
	})

	t.Run("pre-attached attrs are preserved alongside context fields", func(t *testing.T) {
		var buf bytes.Buffer
		inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: LevelTrace})
		logger := slog.New(NewContextHandler(inner)).With("service", "test-service")

		ctx := context.WithValue(context.Background(), ctxkeys.RequestID, "req-abc")

		logger.InfoContext(ctx, "test message")

		assert.True(t, json.Valid(buf.Bytes()), "output must be valid JSON")

		var record map[string]any
		assert.NoError(t, json.Unmarshal(buf.Bytes(), &record))
		assert.Equal(t, "test-service", record["service"])
		assert.Equal(t, "req-abc", record["requestId"])
	})

	t.Run("same group name in WithAttrs and per-call attrs are not deep merged", func(t *testing.T) {
		var buf bytes.Buffer
		inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: LevelTrace})
		logger := slog.New(NewContextHandler(inner)).With(
			slog.Group("metadata", "source", "background"),
		)

		logger.InfoContext(context.Background(), "job started",
			slog.Group("metadata", "jobName", "sync"),
		)

		raw := buf.Bytes()
		assert.True(t, json.Valid(raw), "output must be valid JSON")

		// slog does not deep merge groups — both objects are emitted under the same
		// key as separate entries rather than being combined into one object.
		assert.Contains(t, string(raw), `"source":"background"`)
		assert.Contains(t, string(raw), `"jobName":"sync"`)

		// Standard JSON parsing into a map can only hold one value per key;
		// when duplicate keys are present the last occurrence wins, so the
		// WithAttrs group is shadowed by the per-call group.
		var record map[string]any
		assert.NoError(t, json.Unmarshal(raw, &record))
		metadata, ok := record["metadata"].(map[string]any)
		assert.True(t, ok, "metadata should be a JSON object")
		assert.Contains(t, metadata, "jobName")
		assert.NotContains(t, metadata, "source")
	})
}
