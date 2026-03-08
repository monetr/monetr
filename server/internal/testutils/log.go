package testutils

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/monetr/monetr/server/logging"
)

// TestLogHook captures log records written during a test so that assertions
// can be made about what was logged.
type TestLogHook struct {
	mu      sync.Mutex
	entries []slog.Record
}

func (h *TestLogHook) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *TestLogHook) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Clone the record so attrs are preserved after this call returns.
	clone := r.Clone()
	h.entries = append(h.entries, clone)
	return nil
}

func (h *TestLogHook) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *TestLogHook) WithGroup(name string) slog.Handler        { return h }

func (h *TestLogHook) AllEntries() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]slog.Record, len(h.entries))
	copy(out, h.entries)
	return out
}

func (h *TestLogHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = h.entries[:0]
}

// multiHandler fans out to multiple slog.Handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: next}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: next}
}

type testLogEntry struct {
	log  *slog.Logger
	hook *TestLogHook
}

var testLogs struct {
	lock sync.Mutex
	logs map[string]*testLogEntry
}

func init() {
	testLogs = struct {
		lock sync.Mutex
		logs map[string]*testLogEntry
	}{
		lock: sync.Mutex{},
		logs: map[string]*testLogEntry{},
	}
}

func GetTestLog(t *testing.T) (*slog.Logger, *TestLogHook) {
	testLogs.lock.Lock()
	defer testLogs.lock.Unlock()

	if entry, ok := testLogs.logs[t.Name()]; ok {
		return entry.log, entry.hook
	}

	hook := &TestLogHook{}
	base := logging.NewLoggerWithLevel("trace")

	combined := slog.New(&multiHandler{
		handlers: []slog.Handler{
			base.Handler(),
			hook,
		},
	}).With("test", t.Name())

	t.Cleanup(func() {
		testLogs.lock.Lock()
		defer testLogs.lock.Unlock()
		delete(testLogs.logs, t.Name())
	})

	testLogs.logs[t.Name()] = &testLogEntry{
		log:  combined,
		hook: hook,
	}

	return combined, hook
}

func GetLog(t *testing.T) *slog.Logger {
	log, _ := GetTestLog(t)
	return log
}

func MustHaveLogMessage(t *testing.T, hook *TestLogHook, message string) {
	for _, entry := range hook.AllEntries() {
		if strings.EqualFold(entry.Message, message) {
			return
		}
	}

	t.Fatalf("log message was not sent during test: %s", message)
}
