package logging

import "log/slog"

// LevelTrace is a custom slog level below Debug. logrus had a Trace level
// that slog does not define, so we define it here to preserve the distinction.
const LevelTrace = slog.Level(-8)
