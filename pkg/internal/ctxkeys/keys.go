package ctxkeys

import (
	"context"

	"github.com/sirupsen/logrus"
)

type MonetrContextKey string

const (
	RequestID MonetrContextKey = "requestId"
	UserID    MonetrContextKey = "userId"
	AccountID MonetrContextKey = "accountId"
	LoginID   MonetrContextKey = "loginId"
)

var (
	keys = []MonetrContextKey{
		RequestID,
		UserID,
		AccountID,
		LoginID,
	}
)

// LogrusFieldsFromContext will extract known fields from the Go context in order to better support logging. This allows
// monetr to more easily provide useful context for all of its log entries for every action that occurs.
func LogrusFieldsFromContext(ctx context.Context, existingFields logrus.Fields) logrus.Fields {
	fields := logrus.Fields{}
	for _, key := range keys {
		// If the field is already on the log entry then do not overwrite it.
		if _, ok := existingFields[string(key)]; ok {
			continue
		}

		if value := ctx.Value(key); value != nil {
			fields[string(key)] = value
		}
	}

	return fields
}
