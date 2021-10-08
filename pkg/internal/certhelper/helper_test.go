package certhelper

import (
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewFileCertificateHelper(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		file, err := ioutil.TempFile("", "")
		require.NoError(t, err, "failed to create temp file")

		var counter int32
		var callback Callback = func(path string) error {
			atomic.AddInt32(&counter, 1)
			return nil
		}

		watcher, err := NewFileCertificateHelper(testutils.GetLog(t), []string{
			file.Name(),
		}, callback)
		assert.NoError(t, err, "failed to create helper")

		err = watcher.Start()
		assert.NoError(t, err, "must be able to start")

		assert.EqualValues(t, 0, atomic.LoadInt32(&counter), "change counter should be 0")

		_, err = file.WriteString("test change #1")
		require.NoError(t, err, "failed to write to temp file")

		// No idea how long it takes to poll, but should not take longer than a second for the event to propagate.
		time.Sleep(2 * time.Second)

		assert.EqualValues(t, 1, atomic.LoadInt32(&counter), "change counter should be 1")

		err = os.Remove(file.Name())
		require.NoError(t, err, "failed to delete temp file")

		err = watcher.Stop()
		assert.NoError(t, err, "failed to stop")
	})
}
