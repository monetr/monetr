package crumbs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
)

func IncludeUserInScope(ctx context.Context, accountId uint64) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       strconv.FormatUint(accountId, 10),
				Username: fmt.Sprintf("account:%d", accountId),
			})
			scope.SetTag("accountId", strconv.FormatUint(accountId, 10))
		})
	}
}
