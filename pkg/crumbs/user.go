package crumbs

import (
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
)

func IncludeUserInScope(hub *sentry.Hub, accountId uint64) {
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       strconv.FormatUint(accountId, 10),
			Username: fmt.Sprintf("account:%d", accountId),
		})
		scope.SetTag("accountId", strconv.FormatUint(accountId, 10))
	})
}
