package logging

import (
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"github.com/monetr/monetr/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
)

var (
	_             logrus.Formatter = &stackDriverFormatterWrapper{}
	fieldToLabels                  = []string{
		"accountId",
		"userId",
		"loginId",
		"requestId",
		"jobId",
	}
)

type stackDriverFormatterWrapper struct {
	config config.StackDriverLogging
	inner  logrus.Formatter
	client *logging.Client
	logger *logging.Logger
}

func NewStackDriverFormatterWrapper(inner logrus.Formatter, config config.StackDriverLogging) (logrus.Formatter, error) {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, config.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stackdriver logging client")
	}

	return &stackDriverFormatterWrapper{
		config: config,
		inner:  inner,
		client: client,
		logger: client.Logger(config.LogName),
	}, nil
}

func (s *stackDriverFormatterWrapper) Format(entry *logrus.Entry) ([]byte, error) {
	googleEntry := logging.Entry{
		Timestamp:      entry.Time,
		Severity:       0,
		Payload:        nil,
		Labels:         nil,
		SourceLocation: nil,
	}

	switch entry.Level {
	case logrus.FatalLevel:
		googleEntry.Severity = logging.Alert
	case logrus.ErrorLevel:
		googleEntry.Severity = logging.Error
	case logrus.InfoLevel:
		googleEntry.Severity = logging.Info
	case logrus.DebugLevel, logrus.TraceLevel:
		googleEntry.Severity = logging.Debug
	default:
		// If for some reason we cannot translate the log level then just pass it to the inner.
		return s.inner.Format(entry)
	}

	if entry.Caller != nil {
		googleEntry.SourceLocation = &logpb.LogEntrySourceLocation{
			File:     entry.Caller.File,
			Line:     int64(entry.Caller.Line),
			Function: entry.Caller.Function,
		}
	}

	payload := map[string]interface{}{}
	for key, value := range entry.Data {
		payload[key] = value
	}
	labels := map[string]string{}
	for _, key := range fieldToLabels {
		value, ok := payload[key]
		if !ok {
			continue
		}

		labels[key] = fmt.Sprint(value)
		delete(payload, key)
	}

	payload["msg"] = entry.Message

	s.logger.Log(googleEntry)

	return s.inner.Format(entry)
}
