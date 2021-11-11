package logging

import (
	"github.com/sirupsen/logrus"
)

func duplicateEntry(entry *logrus.Entry) *logrus.Entry {
	duplicate := entry.Dup()
	duplicate.Message = entry.Message
	duplicate.Buffer = entry.Buffer
	duplicate.Level = entry.Level
	duplicate.Caller = entry.Caller
	duplicate.Time = entry.Time

	return duplicate
}
