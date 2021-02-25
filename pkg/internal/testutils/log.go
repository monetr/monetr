package testutils

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/logging"
	"github.com/sirupsen/logrus"
	"testing"
)

func GetLog(t *testing.T) *logrus.Entry {
	return logging.NewLogger().WithField("test", t.Name())
}
