package controller

import (
	"context"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/ctxkeys"
	"github.com/monetr/monetr/server/internal/sentryecho"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
)

const (
	databaseContextKey           = "_monetrDatabase_"
	subscriptionStatusContextKey = "_subscriptionStatus_"
	spanContextKey               = "_spanContext_"
	spanKey                      = "_span_"
	authenticationKey            = "_authentication_"
)

func (c *Controller) databaseRepositoryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
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
				c.Log.ErrorContext(c.getContext(ctx), "failed to begin transaction", "err", err)
				return c.wrapAndReturnError(
					ctx,
					err,
					http.StatusInternalServerError,
					"Internal error, try again in a few moments",
				)
			}

			cleanup = func() {
				panicErr := recover()
				if handlerError != nil || panicErr != nil {
					if err := txn.RollbackContext(c.getContext(ctx)); err != nil {
						c.Log.ErrorContext(c.getContext(ctx), "failed to rollback request", "err", err)
					}
				} else {
					if err := txn.CommitContext(c.getContext(ctx)); err != nil {
						panic(err)
					}
				}
				if panicErr != nil {
					panic(panicErr)
				}
			}

			dbi = txn
		}

		ctx.Set(databaseContextKey, dbi)

		if cleanup != nil {
			defer cleanup()
		}

		handlerError = next(ctx)

		return handlerError
	}
}

func (c *Controller) removeCookieIfPresent(ctx *echo.Context) {
	c.updateAuthenticationCookie(ctx, ClearAuthentication)
}

func (c *Controller) requireActiveSubscriptionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
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

// maybeApiKeyMiddleware allows monetr API keys to be provided via the
// authorization header as a username and password. Where they key ID of the API
// key is the username and the secret is the password. The credentials are
// validated if they are specified. If they are specified and are invalid then
// the request will fail even if a valid session token is present on the
// request.
func (c *Controller) maybeApiKeyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		log := c.getLog(ctx)
		username, password, ok := ctx.Request().BasicAuth()
		if ok {
			now := c.Clock.Now()
			data := map[string]any{
				"source": "key",
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

			repo := c.mustGetUnauthenticatedRepository(ctx)
			keyId, err := models.ParseID[models.ApiKey](username)
			if err != nil {
				log.WarnContext(
					c.getContext(ctx),
					"invalid api key username provided",
				)
				return c.unauthorized(ctx)
			}
			apiKey, err := repo.GetApiKey(c.getContext(ctx), keyId)
			switch {
			case err == nil:
				// Keep going, we found a key for this Id.
			case errors.Is(err, pg.ErrNoRows):
				// There is no key with this Id, the credentials are definitively bad.
				log.WarnContext(
					c.getContext(ctx),
					"invalid api key provided",
					"err", err,
				)
				return c.unauthorized(ctx)
			default:
				// Any other error means we could not determine whether the credentials
				// are valid, the database might be down. Telling the client they are
				// unauthorized would be a lie, and would make an outage look like an
				// authentication problem for anyone using an API key. Fail loudly
				// instead so that this gets reported.
				log.ErrorContext(
					c.getContext(ctx),
					"failed to retrieve api key for authentication",
					"err", err,
				)
				breadcrumbMessage = "Request auth could not be verified"
				return c.wrapPgError(ctx, err, "failed to authenticate api key")
			}

			if !apiKey.Verify(keyId, password) {
				log.WarnContext(
					c.getContext(ctx),
					"invalid api key provided",
					"err", "credential mismatch",
				)
				return c.unauthorized(ctx)
			}

			claims := security.Claims{
				CreatedAt:    c.Clock.Now(),
				EmailAddress: apiKey.CreatedByUser.Login.Email,
				UserId:       apiKey.CreatedBy.String(),
				AccountId:    apiKey.AccountId.String(),
				LoginId:      apiKey.CreatedByUser.LoginId.String(),
				Scope:        security.AuthenticatedScope,
				ReissueCount: 0,
			}

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
			ctx.Set(authenticationKey, claims)

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
		}

		return next(ctx)
	}
}

func (c *Controller) maybeTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) (err error) {
		err = func(ctx *echo.Context) error {
			// If there are already credentials on the scope of the request from
			// something else then do not do anything with the token
			if ctx.Get(authenticationKey) != nil {
				return nil
			}

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

			// If there is still no token then we don't have one. Return nothing.
			if token == "" {
				return nil
			}

			claims, err := c.ClientTokens.Parse(token)
			if err != nil {
				c.updateAuthenticationCookie(ctx, ClearAuthentication)
				crumbs.Error(
					c.getContext(ctx),
					"failed to parse token",
					"authentication",
					map[string]any{
						"error": err,
					},
				)
				log.WarnContext(c.getContext(ctx), "invalid token provided", "err", err)
				return nil
			}

			// If we can pull the hub from the current context, then we want to try to
			// set some of our user data on it so that way we can grab it later if
			// there is an error.
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

			return nil
		}(ctx)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Internal error")
		}

		return next(ctx)
	}
}

// requireAuthentication is an echo middleware that requires that the current
// HTTTP request is authenticated with one of the provided scopes. Any of the
// specified scopes are valid. The scopes can be derived from API keys or from
// the actual token that monetr issues for sessions.
func (c *Controller) requireAuthentication(
	scopes ...security.Scope,
) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
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

// requireLunchFlowEnabledMiddleware is added before the lunch flow API routes
// are registered. This makes sure that if the server is not configured to use
// lunch flow then the routes are not accessible.
func (c *Controller) requireLunchFlowEnabledMiddleware(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		if !c.Configuration.LunchFlow.Enabled {
			return c.notFound(ctx, "Lunch Flow is not enabled on this server")
		}

		return next(ctx)
	}
}
