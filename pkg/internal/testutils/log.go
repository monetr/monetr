package testutils

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/monetr/monetr/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

type testLogEntry struct {
	log  *logrus.Entry
	hook *test.Hook
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

func GetTestLog(t *testing.T) (*logrus.Entry, *test.Hook) {
	testLogs.lock.Lock()
	defer testLogs.lock.Unlock()

	if log, ok := testLogs.logs[t.Name()]; ok {
		return log.log, log.hook
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
		DisableLevelTruncation: false,
		PadLevelText:           true,
		QuoteEmptyFields:       false,
		FieldMap:               nil,
		CallerPrettyfier:       nil,
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

	testHook := test.NewLocal(logger.Logger)
	testLogs.logs[t.Name()] = &testLogEntry{
		log:  logger,
		hook: testHook,
	}

	return logger, testHook
}

func GetLog(t *testing.T) *logrus.Entry {
	log, _ := GetTestLog(t)
	return log
}

func MustHaveLogMessage(t *testing.T, hook *test.Hook, message string) {
	for _, entry := range hook.AllEntries() {
		if strings.EqualFold(entry.Message, message) {
			return
		}
	}

	t.Fatalf("log message was not sent during test: %s", message)
}
