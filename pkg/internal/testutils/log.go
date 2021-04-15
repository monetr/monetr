package testutils

import (
	"github.com/monetrapp/rest-api/pkg/logging"
	"github.com/sirupsen/logrus"
	"testing"
)

func GetLog(t *testing.T) *logrus.Entry {
	return logging.NewLogger().WithField("test", t.Name())
}
