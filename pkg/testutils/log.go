package testutils

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func GetLog(t *testing.T) *logrus.Entry {
	logger := logrus.New()
	log := logger.WithField("test", t.Name())
	return log
}
