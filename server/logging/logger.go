package logging

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/acaloiaro/neoq/logging"
	"github.com/monetr/monetr/server/config"
	"github.com/sirupsen/logrus"
)

func NewLoggerWithConfig(configuration config.Logging) *logrus.Entry {
	logger := logrus.New()
	logger.Out = os.Stderr

	level, err := logrus.ParseLevel(configuration.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	switch strings.ToLower(configuration.Format) {
	default:
		fallthrough
	case "text":
		logger.Formatter = &logrus.TextFormatter{
			ForceColors:               false,
			DisableColors:             false,
			ForceQuote:                true,
			DisableQuote:              false,
			EnvironmentOverrideColors: false,
			DisableTimestamp:          true,
			FullTimestamp:             false,
			TimestampFormat:           "",
			DisableSorting:            false,
			SortingFunc: func(input []string) {
				if len(input) == 0 {
					return
				}
				keys := make([]string, 0, len(input)-1)
				for _, key := range input {
					if key == "msg" {
						continue
					}

					keys = append(keys, key)
				}
				sort.Strings(keys)
				keys = append(keys, "msg")
				copy(input, keys)
			},
			DisableLevelTruncation: false,
			PadLevelText:           true,
			QuoteEmptyFields:       false,
			FieldMap:               nil,
			CallerPrettyfier:       nil,
		}
	case "json":
		formatter := &logrus.JSONFormatter{
			TimestampFormat:   time.RFC3339,
			DisableTimestamp:  false,
			DisableHTMLEscape: false,
			DataKey:           "",
			FieldMap:          logrus.FieldMap{},
			CallerPrettyfier:  nil,
			PrettyPrint:       false,
		}

		// If we are using stackdriver, then use the `message` field name instead. Stackdriver will automatically detect
		// this field in the json payload of a log entry.
		if configuration.StackDriver.Enabled {
			formatter.FieldMap = logrus.FieldMap{
				logrus.FieldKeyMsg: "message",
			}
		}

		logger.Formatter = formatter
	}

	// If stack driver is enabled then we need to perform some mutations around the formatter before the object is
	// printed. This wrapper must come ahead of the context wrapper because of how the chain of wrappers is built.
	// This makes sure that the context fields are properly placed on the stackdriver formatted log message.
	if configuration.StackDriver.Enabled {
		formatter, err := NewStackDriverFormatterWrapper(logger.Formatter)
		if err != nil {
			logger.WithError(err).Errorf("failed to create stack driver wrapper")
			return logrus.NewEntry(logger)
		}

		logger.Formatter = formatter
	}

	logger.Formatter = NewContextFormatterWrapper(logger.Formatter)

	return logrus.NewEntry(logger)
}

func NewLogger() *logrus.Entry {
	return NewLoggerWithLevel(logrus.InfoLevel.String())
}

func NewLoggerWithLevel(levelString string) *logrus.Entry {
	return NewLoggerWithConfig(config.Logging{
		Level: levelString,
	})
}

var (
	_ logging.Logger = &logrusWrapper{}
)

type logrusWrapper struct {
	log *logrus.Entry
}

func NewLogrusWrapper(log *logrus.Entry) logging.Logger {
	return &logrusWrapper{
		log: log,
	}
}

// Debug implements logging.Logger.
func (l *logrusWrapper) Debug(msg string, args ...any) {
	l.write(logrus.DebugLevel, msg, args...)
}

// Error implements logging.Logger.
func (l *logrusWrapper) Error(msg string, args ...any) {
	l.write(logrus.ErrorLevel, msg, args...)
}

// Info implements logging.Logger.
func (l *logrusWrapper) Info(msg string, args ...any) {
	l.write(logrus.InfoLevel, msg, args...)
}

func (l *logrusWrapper) write(level logrus.Level, msg string, args ...any) {
	fields := logrus.Fields{}
	for i, arg := range args {
		if i == 0 && len(args)%2 == 0 {
			// If this is the first arg and there are an even number of fields, assume its key value pairs and skip to the
			// next one.
			continue
		} else if len(args)%2 == 0 && i%2 == 0 {
			// If this is not the first one, and there are an even number of fields, and we are on an even arg. Then assume
			// the previous item was a key and this is the value.
			fields[fmt.Sprint(args[i-1])] = arg
		} else if len(args)%2 == 0 && i%2 != 0 {
			// If there are an even number of fields but this is not the even field. Then keep moving this is a key not a
			// value.
			continue
		} else {
			// If there arent an even number of fields then log based on the index instead.
			fields[strconv.FormatInt(int64(i), 10)] = arg
		}
	}

	l.log.WithFields(fields).Log(level, msg)
}
