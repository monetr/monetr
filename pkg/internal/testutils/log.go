package testutils

import (
	"bytes"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"sort"
	"sync"
	"testing"
)

var testLogs struct {
	lock sync.Mutex
	logs map[string]*logrus.Entry
}

func init() {
	testLogs = struct {
		lock sync.Mutex
		logs map[string]*logrus.Entry
	}{
		lock: sync.Mutex{},
		logs: map[string]*logrus.Entry{},
	}
}

func GetLog(t *testing.T) *logrus.Entry {
	testLogs.lock.Lock()
	defer testLogs.lock.Unlock()

	if log, ok := testLogs.logs[t.Name()]; ok {
		return log
	}

	logger := logging.NewLogger()
	logger.Logger.Level = logrus.TraceLevel

	output := bytes.NewBuffer(nil)
	logger.Logger.Out = output
	logger.Logger.Formatter = &logrus.TextFormatter{
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              true,
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
		DisableLevelTruncation:    false,
		PadLevelText:              true,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	}

	t.Cleanup(func() {
		testLogs.lock.Lock()
		defer testLogs.lock.Unlock()

		delete(testLogs.logs, t.Name())

		if t.Failed() {
			_, err := os.Stderr.Write(output.Bytes())
			require.NoError(t, err, "must write failed logs")
		}
	})

	logger = logger.WithField("test", t.Name())

	testLogs.logs[t.Name()] = logger

	return logger
}
