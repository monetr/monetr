package testutils

import (
	"bytes"
	"github.com/monetr/rest-api/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
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
