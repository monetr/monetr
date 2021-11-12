package logging

import (
	"github.com/sirupsen/logrus"
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

	// Stackdriver log levels documented here:
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
	levelsToStackdriver = map[logrus.Level]string{
		logrus.FatalLevel: "EMERGENCY",
		logrus.ErrorLevel: "ERROR",
		logrus.InfoLevel:  "NOTICE",
		logrus.DebugLevel: "INFO",
		logrus.TraceLevel: "DEBUG",
	}
)

type stackDriverFormatterWrapper struct {
	inner logrus.Formatter
}

func NewStackDriverFormatterWrapper(inner logrus.Formatter) (logrus.Formatter, error) {
	return &stackDriverFormatterWrapper{
		inner: inner,
	}, nil
}

func (s *stackDriverFormatterWrapper) Format(entry *logrus.Entry) ([]byte, error) {
	// This entire format wrapper moves some log entry data around for stackdriver.
	// These behaviors are documented here:
	// https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields
	duplicate := duplicateEntry(entry, logrus.Fields{
		"severity": levelsToStackdriver[entry.Level],
		"message":  entry.Message,
	})

	// Build our labels map.
	labels := map[string]interface{}{}
	for _, label := range fieldToLabels {
		value, ok := duplicate.Data[label]
		if !ok {
			continue
		}

		labels[label] = value
	}

	if len(labels) > 0 {
		// Remove these fields from the data map since we are moving them to the labels map.
		for label := range labels {
			delete(duplicate.Data, label)
		}
		duplicate.Data["logging.googleapis.com/labels"] = labels
	}

	return s.inner.Format(duplicate)
}
