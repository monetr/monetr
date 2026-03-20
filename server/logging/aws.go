package logging

import (
	"context"
	"fmt"
	"log/slog"

	awslogging "github.com/aws/smithy-go/logging"
)

var (
	_ awslogging.Logger = &AWSLogger{}
)

// AWSLogger adapts an *slog.Logger to the smithy-go logging.Logger interface
// used by the AWS SDK v2, ensuring SDK log output is structured JSON (or
// whatever format the slog handler produces).
type AWSLogger struct {
	log *slog.Logger
}

func NewAWSLogger(log *slog.Logger) *AWSLogger {
	return &AWSLogger{
		log: log.WithGroup("aws"),
	}
}

func (a *AWSLogger) Logf(
	classification awslogging.Classification,
	format string,
	v ...any,
) {
	var level slog.Level
	switch classification {
	case awslogging.Warn:
		level = slog.LevelWarn
	default:
		level = slog.LevelDebug
	}
	a.log.Log(context.Background(), level, fmt.Sprintf(format, v...))
}
