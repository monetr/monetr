package crumbs

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
)

func IncludeUserInScope(ctx context.Context, accountId fmt.Stringer) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       accountId.String(),
				Username: accountId.String(),
			})
			scope.SetTag("accountId", accountId.String())
		})
	}
}
