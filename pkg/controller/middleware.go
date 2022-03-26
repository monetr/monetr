package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/getsentry/sentry-go"
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/internal/ctxkeys"
	"github.com/monetr/monetr/pkg/repository"
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

func (c *Controller) setupRepositoryMiddleware(ctx iris.Context) {
	var cleanup func()
	var dbi pg.DBI
	switch ctx.Method() {
	case "GET", "OPTIONS":
		dbi = c.db
	case "POST", "PUT", "DELETE":
		txn, err := c.db.BeginContext(c.getContext(ctx))
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to begin transaction")
			return
		}

		cleanup = func() {
			// TODO (elliotcourant) Add proper logging here that the request has failed
			//  and we are rolling back the transaction.
			if ctx.GetErr() != nil {
				if err := txn.RollbackContext(c.getContext(ctx)); err != nil {
					// Rollback
					c.log.WithError(err).Errorf("failed to rollback request")
				}
			} else {
				if err = txn.CommitContext(c.getContext(ctx)); err != nil {
					// failed to commit
					fmt.Println(err)
				}
			}
		}

		dbi = txn
	}

	ctx.Values().Set(databaseContextKey, dbi)

	ctx.Next()

	if cleanup != nil {
		cleanup()
	}

	ctx.Next()
}

func (c *Controller) removeCookieIfPresent(ctx iris.Context) {
	ctx.RemoveCookie(c.configuration.Server.Cookies.Name)
}

func (c *Controller) authenticateUser(ctx iris.Context) (err error) {
	now := time.Now()
	var token string

	data := map[string]interface{}{
		"source": "none",
	}

	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		defer func() {
			var message string
			if err == nil {
				message = "Token is valid"
				data["accountId"] = c.mustGetAccountId(ctx)
				data["userId"] = c.mustGetUserId(ctx)
			} else {
				message = "Request did not have valid Token"
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

	{ // Read the token from the request.
		// We can allocate a max capacity of 2 right away because we know (at least at the time of writing this) that we
		// will have _at most_ 2 cookie options.
		cookieOptions := make([]iris.CookieOption, 0, 2)
		cookieOptions = append(cookieOptions, iris.CookieHTTPOnly(true))

		// If the server is configured to use secure cookies then add that to the options.
		if c.configuration.Server.Cookies.Secure {
			cookieOptions = append(cookieOptions, iris.CookieSecure)
		}

		// Try to retrieve the cookie from the request with the options.
		if token = ctx.GetCookie(
			c.configuration.Server.Cookies.Name,
			cookieOptions...,
		); token != "" {
			data["source"] = "cookie"
		}
	}

	if token == "" {
		return errors.Errorf("token must be provided")
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
		c.removeCookieIfPresent(ctx)
		return errors.Wrap(err, "failed to validate token")
	}

	if !result.Valid {
		c.removeCookieIfPresent(ctx)
		return errors.Errorf("token is not valid")
	}

	// If we can pull the hub from the current context, then we want to try to set some of our user data on it so that
	// way we can grab it later if there is an error.
	if hub := sentryiris.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetUser(sentry.User{
			ID:        strconv.FormatUint(claims.AccountId, 10),
			Username:  fmt.Sprintf("account:%d", claims.AccountId),
			IPAddress: ctx.GetHeader("X-Forwarded-For"),
		})
		hub.Scope().SetTag("userId", strconv.FormatUint(claims.UserId, 10))
		hub.Scope().SetTag("accountId", strconv.FormatUint(claims.AccountId, 10))
		hub.Scope().SetTag("loginId", strconv.FormatUint(claims.LoginId, 10))
	}

	ctx.Values().Set(accountIdContextKey, claims.AccountId)
	ctx.Values().Set(userIdContextKey, claims.UserId)
	ctx.Values().Set(loginIdContextKey, claims.LoginId)

	{ // Add some basic values onto our context for logging later on.
		spanContext := ctx.Values().Get(spanContextKey).(context.Context)
		spanContext = context.WithValue(spanContext, ctxkeys.AccountID, claims.AccountId)
		spanContext = context.WithValue(spanContext, ctxkeys.UserID, claims.UserId)
		spanContext = context.WithValue(spanContext, ctxkeys.LoginID, claims.LoginId)
		ctx.Values().Set(spanContextKey, spanContext)
	}

	return nil
}

func (c *Controller) authenticationMiddleware(ctx iris.Context) {
	if err := c.authenticateUser(ctx); err != nil {
		c.updateAuthenticationCookie(ctx, ClearAuthentication)
		ctx.SetErr(err)
		ctx.StatusCode(http.StatusForbidden)
		ctx.StopExecution()
		return
	}

	ctx.Next()
}

func (c *Controller) requireActiveSubscriptionMiddleware(ctx iris.Context) {
	accountId := c.mustGetAccountId(ctx)

	active, err := c.paywall.GetSubscriptionIsActive(c.getContext(ctx), accountId)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to validate subscription is active")
		return
	}

	if !active {
		c.getSpan(ctx).Status = sentry.SpanStatusPermissionDenied
		c.returnError(ctx, http.StatusPaymentRequired, "subscription is not active")
		return
	}

	ctx.Next()
}

func (c *Controller) loggingMiddleware(ctx iris.Context) {
	ctx.Next()

	if err := ctx.GetErr(); err != nil {
		c.getLog(ctx).WithError(err).Errorf("%s", ctx.GetErr().Error())
	}
}

func (c *Controller) mustGetDatabase(ctx iris.Context) pg.DBI {
	txn, ok := ctx.Values().Get(databaseContextKey).(*pg.Tx)
	if !ok {
		panic("no database on context")
	}

	return txn
}

// mustGetSecurityRepository is used to retrieve/create a repository interface that can interact with more security
// sensitive parts of the data layer. This interface is not specific to a single tenant. If the interface cannot be
// created due then this method will panic.
func (c *Controller) mustGetSecurityRepository(ctx iris.Context) repository.SecurityRepository {
	db, ok := ctx.Values().Get(databaseContextKey).(pg.DBI)
	if !ok {
		panic("failed to retrieve database object from controller context")
	}

	return repository.NewSecurityRepository(db)
}

func (c *Controller) getUnauthenticatedRepository(ctx iris.Context) (repository.UnauthenticatedRepository, error) {
	txn, ok := ctx.Values().Get(databaseContextKey).(*pg.Tx)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewUnauthenticatedRepository(txn), nil
}

func (c *Controller) mustGetUnauthenticatedRepository(ctx iris.Context) repository.UnauthenticatedRepository {
	repo, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		panic(err)
	}

	return repo
}

func (c *Controller) mustGetUserId(ctx iris.Context) uint64 {
	userId := ctx.Values().GetUint64Default(userIdContextKey, 0)
	if userId == 0 {
		panic("unauthorized")
	}

	return userId
}

func (c *Controller) mustGetAccountId(ctx iris.Context) uint64 {
	accountId := ctx.Values().GetUint64Default(accountIdContextKey, 0)
	if accountId == 0 {
		panic("unauthorized")
	}

	return accountId
}

func (c *Controller) getAuthenticatedRepository(ctx iris.Context) (repository.Repository, error) {
	loginId := ctx.Values().GetUint64Default(loginIdContextKey, 0)
	if loginId == 0 {
		return nil, errors.Errorf("not authorized")
	}

	userId := ctx.Values().GetUint64Default(userIdContextKey, 0)
	if userId == 0 {
		return nil, errors.Errorf("you are not authenticated to an account")
	}

	accountId := ctx.Values().GetUint64Default(accountIdContextKey, 0)
	if accountId == 0 {
		return nil, errors.Errorf("you are not authenticated to an account")
	}

	txn, ok := ctx.Values().Get(databaseContextKey).(pg.DBI)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewRepositoryFromSession(userId, accountId, txn), nil
}

func (c *Controller) mustGetAuthenticatedRepository(ctx iris.Context) repository.Repository {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		panic("unauthorized")
	}

	return repo
}
