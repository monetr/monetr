package testutils

import (
	"log/slog"
	"testing"

	"github.com/monetr/monetr/server/internal/testutils/testlog"
)

// The actual log helpers live in server/internal/testutils/testlog so that
// packages testutils itself depends on (such as server/migrations) can use
// them without us tripping over an import cycle. Everything in this file is
// a thin wrapper that preserves the historic testutils.* surface.

type TestLogHook = testlog.TestLogHook

func GetTestLog(t *testing.T) (*slog.Logger, *TestLogHook) {
	return testlog.GetTestLog(t)
}

func GetLog(t *testing.T) *slog.Logger {
	return testlog.GetLog(t)
}

func MustHaveLogMessage(t *testing.T, hook *TestLogHook, message string) {
	testlog.MustHaveLogMessage(t, hook, message)
}
