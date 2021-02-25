package logging

import (
	"github.com/sirupsen/logrus"
	"os"
	"sort"
)

func NewLogger() *logrus.Entry {
	logger := logrus.New()
	if os.Getenv("CI") == "" {
		logger.SetLevel(logrus.TraceLevel)
	} else {
		logger.SetLevel(logrus.FatalLevel)
	}

	logger.Formatter = &logrus.TextFormatter{
		ForceColors:               false,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              true,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          true,
		FullTimestamp:             false,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc: func(input []string) {
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
