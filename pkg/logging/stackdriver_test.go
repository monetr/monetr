package logging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestStackDriverFormatterWrapper(t *testing.T) {
	log := logrus.NewEntry(logrus.StandardLogger())
	log.Logger.SetLevel(logrus.TraceLevel)

	log = log.WithField("accountId", uint64(1234))

	formatter, err := NewStackDriverFormatterWrapper(&logrus.JSONFormatter{})
	assert.NoError(t, err, "must not return an error just creating the wrapper")
	assert.NotNil(t, formatter, "returned formatter must not be nil")

	log.Message = "I am a log message"
	log.Level = logrus.InfoLevel
	log.Time = time.Now()

	result, err := formatter.Format(log)
	assert.NoError(t, err, "should format log successfully")
	assert.True(t, json.Valid(result), "result must be valid json")

	var object map[string]interface{}
	assert.NoError(t, json.Unmarshal(result, &object), "must unmarshal log entry successfully")

	assert.Contains(t, object, "severity", "must contain the severity field for stackdriver")
	assert.Contains(t, object, "logging.googleapis.com/labels", "must contain the labels field for stackdriver")
}
