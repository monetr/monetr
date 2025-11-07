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
)

const (
	databaseContextKey           = "_monetrDatabase_"
	subscriptionStatusContextKey = "_subscriptionStatus_"
	spanContextKey               = "_spanContext_"
	spanKey                      = "_span_"
	authenticationKey            = "_authentication_"
)

func (c *Controller) databaseRepositoryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var cleanup func()
		var dbi pg.DBI
		var handlerError error
		switch strings.ToUpper(ctx.Request().Method) {
		case "GET", "OPTIONS":
			dbi = c.DB
		case "POST":
			// Some endpoints need a POST even though they do not require data access.
			// This is a short term fix. (Hopefully)
			if strings.HasSuffix(ctx.Path(), "/icons/search") {
				dbi = c.DB
				break
			}
			fallthrough
		case "PATCH", "PUT", "DELETE":
			txn, err := c.DB.BeginContext(c.getContext(ctx))
			if err != nil {
				c.Log.WithError(err).Errorf("failed to begin transaction")
				return c.wrapAndReturnError(
					ctx,
					err,
					http.StatusInternalServerError,
					"Internal error, try again in a few moments",
				)
			}

			cleanup = func() {
				// TODO (elliotcourant) Add proper logging here that the request has failed
				//  and we are rolling back the transaction.
				if handlerError != nil {
					if err := txn.RollbackContext(c.getContext(ctx)); err != nil {
						// Rollback
						c.Log.WithError(err).Errorf("failed to rollback request")
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

func (c *Controller) requireActiveSubscriptionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if !c.Configuration.Stripe.IsBillingEnabled() {
			return next(ctx)
		}

		accountId := c.mustGetAccountId(ctx)

		active, err := c.Billing.GetSubscriptionIsActive(c.getContext(ctx), accountId)
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

func (c *Controller) maybeTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		err = func(ctx echo.Context) error {
			now := c.Clock.Now()
			log := c.getLog(ctx)
			var token string
			data := map[string]any{
				"source": "none",
			}
			breadcrumbMessage := "Request did not have valid auth"

			if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
				defer func() {
					hub.AddBreadcrumb(&sentry.Breadcrumb{
						Type:      "debug",
						Category:  "authentication",
						Message:   breadcrumbMessage,
						Data:      data,
						Level:     sentry.LevelDebug,
						Timestamp: now,
					}, nil)
				}()
			}

			{ // Try to retrieve the cookie from the request with the options.
				if tokenCookie, err := ctx.Cookie(
					c.Configuration.Server.Cookies.Name,
				); err == nil && tokenCookie.Value != "" {
					token = tokenCookie.Value
					data["source"] = "cookie"
				}
			}

			if token == "" {
				if token = ctx.Request().Header.Get(c.Configuration.Server.Cookies.Name); token != "" {
					data["source"] = "header"
				}
			}

			// If there is still no token then we don't have one. Return nothing.
			if token == "" {
				return nil
			}

			claims, err := c.ClientTokens.Parse(token)
			if err != nil {
				c.updateAuthenticationCookie(ctx, ClearAuthentication)
				crumbs.Error(c.getContext(ctx), "failed to parse token", "authentication", map[string]any{
					"error": err,
				})
				log.WithError(err).Warn("invalid token provided")
				return nil
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

			// Store the authentication claims on the request context so we can use it
			// later.
			ctx.Set(authenticationKey, *claims)

			{ // Add some basic values onto our context for logging later on.
				spanContext := ctx.Get(spanContextKey).(context.Context)
				spanContext = context.WithValue(spanContext, ctxkeys.AccountID, claims.AccountId)
				spanContext = context.WithValue(spanContext, ctxkeys.UserID, claims.UserId)
				spanContext = context.WithValue(spanContext, ctxkeys.LoginID, claims.LoginId)
				ctx.Set(spanContextKey, spanContext)
			}

			breadcrumbMessage = "Auth is valid"
			data["accountId"] = claims.AccountId
			data["userId"] = claims.UserId
			data["loginId"] = claims.LoginId

			return nil
		}(ctx)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Internal error")
		}

		return next(ctx)
	}
}

// requireToken is an echo middleware that requires that the current HTTTP
// request has a token with one of the provided scopes. Any of the specified
// scopes are valid.
func (c *Controller) requireToken(scopes ...security.Scope) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Read the claims off of the current request context. If the claims are
			// present and valid then this will proceed. If not then this will return
			// an error.
			claims, err := c.getClaims(ctx)
			if err != nil {
				// If there is an error then we are not fully authenticated.
				return c.unauthorizedError(ctx, err)
			}

			// Now validate that the claims have at least one of the required scopes
			// that was specified. If we do not have at least one of the required
			// scopes then the client is not authorized to access the current
			// endpoint.
			if err := claims.RequireScope(scopes...); err != nil {
				return c.unauthorizedError(ctx, err)
			}

			// Everything looks good, run the next middleware or the actual controller
			// function.
			return next(ctx)
		}
	}
}
