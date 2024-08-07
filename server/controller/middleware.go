package controller

import (
	"context"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/ctxkeys"
	"github.com/monetr/monetr/server/internal/sentryecho"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
)

const (
	databaseContextKey           = "_monetrDatabase_"
	accountIdContextKey          = "_accountId_"
	userIdContextKey             = "_userId_"
	loginIdContextKey            = "_loginId_"
	subscriptionStatusContextKey = "_subscriptionStatus_"
	spanContextKey               = "_spanContext_"
	spanKey                      = "_span_"
)

func (c *Controller) databaseRepositoryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var cleanup func()
		var dbi pg.DBI
		var handlerError error
		switch strings.ToUpper(ctx.Request().Method) {
		case "GET", "OPTIONS":
			dbi = c.db
		case "POST":
			// Some endpoints need a POST even though they do not require data access.
			// This is a short term fix. (Hopefully)
			if strings.HasSuffix(ctx.Path(), "/icons/search") {
				dbi = c.db
				break
			}
			fallthrough
		case "PUT", "DELETE":
			txn, err := c.db.BeginContext(c.getContext(ctx))
			if err != nil {
				return c.wrapAndReturnError(
					ctx,
					err,
					http.StatusInternalServerError,
					"failed to begin transaction",
				)
			}

			cleanup = func() {
				// TODO (elliotcourant) Add proper logging here that the request has failed
				//  and we are rolling back the transaction.
				if handlerError != nil {
					if err := txn.RollbackContext(c.getContext(ctx)); err != nil {
						// Rollback
						c.log.WithError(err).Errorf("failed to rollback request")
					}
				} else {
					if err = txn.CommitContext(c.getContext(ctx)); err != nil {
						// failed to commit
						panic(err)
					}
				}
			}

			dbi = txn
		}

		ctx.Set(databaseContextKey, dbi)

		handlerError = next(ctx)

		if cleanup != nil {
			cleanup()
		}

		return handlerError
	}
}

func (c *Controller) removeCookieIfPresent(ctx echo.Context) {
	c.updateAuthenticationCookie(ctx, ClearAuthentication)
}

func (c *Controller) authenticateUser(ctx echo.Context) (err error) {
	now := c.clock.Now()
	log := c.getLog(ctx)
	var token string
	data := map[string]interface{}{
		"source": "none",
	}

	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		defer func() {
			var message string
			if err == nil {
				message = "Auth is valid"
				data["accountId"] = c.mustGetAccountId(ctx)
				data["userId"] = c.mustGetUserId(ctx)
			} else {
				message = "Request did not have valid auth"
			}

			hub.AddBreadcrumb(&sentry.Breadcrumb{
				Type:      "debug",
				Category:  "authentication",
				Message:   message,
				Data:      data,
				Level:     sentry.LevelDebug,
				Timestamp: now,
			}, nil)
		}()
	}

	{ // Try to retrieve the cookie from the request with the options.
		if tokenCookie, err := ctx.Cookie(
			c.configuration.Server.Cookies.Name,
		); err == nil && tokenCookie.Value != "" {
			token = tokenCookie.Value
			data["source"] = "cookie"
		}
	}

	if token == "" {
		if token = ctx.Request().Header.Get(c.configuration.Server.Cookies.Name); token != "" {
			data["source"] = "header"
		}
	}

	if token == "" {
		return errors.New("token must be provided")
	}

	claims, err := c.clientTokens.Parse(security.AuthenticatedAudience, token)
	if err != nil {
		c.updateAuthenticationCookie(ctx, ClearAuthentication)
		// Don't return the JWT error to the client, but throw it in Sentry so it can still be used for debugging.
		crumbs.Error(c.getContext(ctx), "failed to validate token", "authentication", map[string]interface{}{
			"error": err,
		})
		log.WithError(err).Warn("invalid token provided")
		return errors.New("token is not valid")
	}

	// If we can pull the hub from the current context, then we want to try to set some of our user data on it so that
	// way we can grab it later if there is an error.
	if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetUser(sentry.User{
			ID:        claims.AccountId,
			Username:  claims.AccountId,
			IPAddress: util.GetForwardedFor(ctx),
			Data: map[string]string{
				"userId":  claims.UserId,
				"loginId": claims.LoginId,
			},
		})
		hub.Scope().SetTag("userId", claims.UserId)
		hub.Scope().SetTag("accountId", claims.AccountId)
		hub.Scope().SetTag("loginId", claims.LoginId)
	}

	ctx.Set(accountIdContextKey, claims.AccountId)
	ctx.Set(userIdContextKey, claims.UserId)
	ctx.Set(loginIdContextKey, claims.LoginId)

	{ // Add some basic values onto our context for logging later on.
		spanContext := ctx.Get(spanContextKey).(context.Context)
		spanContext = context.WithValue(spanContext, ctxkeys.AccountID, claims.AccountId)
		spanContext = context.WithValue(spanContext, ctxkeys.UserID, claims.UserId)
		spanContext = context.WithValue(spanContext, ctxkeys.LoginID, claims.LoginId)
		ctx.Set(spanContextKey, spanContext)
	}

	return nil
}

func (c *Controller) authenticationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if err := c.authenticateUser(ctx); err != nil {
			c.updateAuthenticationCookie(ctx, ClearAuthentication)
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		return next(ctx)
	}
}

func (c *Controller) requireActiveSubscriptionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if !c.configuration.Stripe.IsBillingEnabled() {
			return next(ctx)
		}

		accountId := c.mustGetAccountId(ctx)

		active, err := c.paywall.GetSubscriptionIsActive(c.getContext(ctx), accountId)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to validate subscription is active")
		}

		if !active {
			c.getSpan(ctx).Status = sentry.SpanStatusPermissionDenied
			return c.returnError(ctx, http.StatusPaymentRequired, "subscription is not active")
		}

		return next(ctx)
	}
}
