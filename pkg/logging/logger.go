package logging

import (
	"github.com/sirupsen/logrus"
	"os"
	"sort"
)

func NewLogger() *logrus.Entry {
	return NewLoggerWithLevel(logrus.FatalLevel.String())
}

func NewLoggerWithLevel(levelString string) *logrus.Entry {
	logger := logrus.New()
	if os.Getenv("CI") != "" {
		logger.SetLevel(logrus.FatalLevel)
	}

	level, err := logrus.ParseLevel(levelString)
	if err != nil {
		level = logrus.InfoLevel
	}

	logger.SetLevel(level)

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
	return logrus.NewEntry(logger)
}
