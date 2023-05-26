package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/ctxkeys"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/util"
	"github.com/pkg/errors"
)

const (
	databaseContextKey           = "_harderDatabase_"
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
	now := time.Now()
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

	var claims MonetrClaims
	result, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(c.configuration.JWT.LoginJwtSecret), nil
	})
	if err != nil {
		c.updateAuthenticationCookie(ctx, ClearAuthentication)
		// Don't return the JWT error to the client, but throw it in Sentry so it can still be used for debugging.
		crumbs.Error(c.getContext(ctx), "failed to validate token", "authentication", map[string]interface{}{
			"error": err,
		})
		return errors.New("token is not valid")
	}

	if !result.Valid {
		c.removeCookieIfPresent(ctx)
		return errors.New("token is not valid")
	}

	// If we can pull the hub from the current context, then we want to try to set some of our user data on it so that
	// way we can grab it later if there is an error.
	if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetUser(sentry.User{
			ID:        strconv.FormatUint(claims.AccountId, 10),
			Username:  fmt.Sprintf("account:%d", claims.AccountId),
			IPAddress: util.GetForwardedFor(ctx),
		})
		hub.Scope().SetTag("userId", strconv.FormatUint(claims.UserId, 10))
		hub.Scope().SetTag("accountId", strconv.FormatUint(claims.AccountId, 10))
		hub.Scope().SetTag("loginId", strconv.FormatUint(claims.LoginId, 10))
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

func (c *Controller) mustGetDatabase(ctx echo.Context) pg.DBI {
	txn, ok := ctx.Get(databaseContextKey).(*pg.Tx)
	if !ok {
		panic("no database on context")
	}

	return txn
}

// mustGetSecurityRepository is used to retrieve/create a repository interface that can interact with more security
// sensitive parts of the data layer. This interface is not specific to a single tenant. If the interface cannot be
// created due then this method will panic.
func (c *Controller) mustGetSecurityRepository(ctx echo.Context) repository.SecurityRepository {
	db, ok := ctx.Get(databaseContextKey).(pg.DBI)
	if !ok {
		panic("failed to retrieve database object from controller context")
	}

	return repository.NewSecurityRepository(db)
}

func (c *Controller) getUnauthenticatedRepository(ctx echo.Context) (repository.UnauthenticatedRepository, error) {
	txn, ok := ctx.Get(databaseContextKey).(*pg.Tx)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewUnauthenticatedRepository(txn), nil
}

func (c *Controller) mustGetUnauthenticatedRepository(ctx echo.Context) repository.UnauthenticatedRepository {
	repo, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		panic(err)
	}

	return repo
}

func (c *Controller) mustGetUserId(ctx echo.Context) uint64 {
	userId, ok := ctx.Get(userIdContextKey).(uint64)
	if userId == 0 || !ok {
		panic("unauthorized")
	}

	return userId
}

func (c *Controller) mustGetAccountId(ctx echo.Context) uint64 {
	accountId, ok := ctx.Get(accountIdContextKey).(uint64)
	if accountId == 0 || !ok {
		panic("unauthorized")
	}

	return accountId
}

func (c *Controller) getAuthenticatedRepository(ctx echo.Context) (repository.Repository, error) {
	loginId, ok := ctx.Get(loginIdContextKey).(uint64)
	if loginId == 0 || !ok {
		return nil, errors.Errorf("not authorized")
	}

	userId, ok := ctx.Get(userIdContextKey).(uint64)
	if userId == 0 || !ok {
		return nil, errors.Errorf("you are not authenticated to an account")
	}

	accountId, ok := ctx.Get(accountIdContextKey).(uint64)
	if accountId == 0 || !ok {
		return nil, errors.Errorf("you are not authenticated to an account")
	}

	txn, ok := ctx.Get(databaseContextKey).(pg.DBI)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewRepositoryFromSession(userId, accountId, txn), nil
}

func (c *Controller) mustGetAuthenticatedRepository(ctx echo.Context) repository.Repository {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		panic("unauthorized")
	}

	return repo
}
