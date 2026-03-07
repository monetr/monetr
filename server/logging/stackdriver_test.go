package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackDriverHandler(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: LevelTrace,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})
	handler := NewStackDriverHandler(inner)
	logger := slog.New(handler).With("accountId", uint64(1234))

	logger.InfoContext(context.Background(), "I am a log message")

	assert.True(t, json.Valid(buf.Bytes()), "result must be valid json")

	var object map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &object), "must unmarshal log entry successfully")

	assert.Contains(t, object, "severity", "must contain the severity field for stackdriver")
	assert.Contains(t, object, "logging.googleapis.com/labels", "must contain the labels field for stackdriver")
}
