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
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/labstack/echo/v4"
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

func (c *Controller) setupRepositoryMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		switch strings.ToUpper(ctx.Request().Method) {
		case "POST", "PUT", "DELETE":
			// Any methods that can cause data to be written should be performed from within a transaction.
			return c.db.RunInTransaction(c.getContext(ctx), func(txn *pg.Tx) error {
				ctx.Set(databaseContextKey, txn)
				return next(ctx)
			})
		default:
			// Otherwise, the request is performed outside a transaction when data is only being read.
			ctx.Set(databaseContextKey, c.db)
			return next(ctx)
		}
	}
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
		cookie, err := ctx.Cookie(c.configuration.Server.Cookies.Name)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve authentication cookie")
		}
		if token = cookie.Value; token != "" {
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
		return errors.Wrap(err, "failed to validate token")
	}

	if !result.Valid {
		return errors.Errorf("token is not valid")
	}

	// If we can pull the hub from the current context, then we want to try to set some of our user data on it so that
	// way we can grab it later if there is an error.
	if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetUser(sentry.User{
			ID:        strconv.FormatUint(claims.AccountId, 10),
			Username:  fmt.Sprintf("account:%d", claims.AccountId),
			IPAddress: ctx.Request().Header.Get("X-Forwarded-For"),
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

func (c *Controller) authenticationMiddleware(ctx echo.Context) {
	if err := c.authenticateUser(ctx); err != nil {
		c.updateAuthenticationCookie(ctx, ClearAuthentication)
		ctx.SetErr(err)
		ctx.StatusCode(http.StatusForbidden)
		ctx.StopExecution()
		return
	}

	ctx.Next()
}

func (c *Controller) requireActiveSubscriptionMiddleware(ctx echo.Context) {
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

func (c *Controller) loggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				c.getLog(ctx).Errorf("PANIC: %+v", err)
				panic(err)
			}
		}()

		if err := next(ctx); err != nil {
			c.getLog(ctx).WithError(err).Errorf("%s", err.Error())
			return err
		}

		return nil
	}
}

func (c *Controller) mustGetDatabase(ctx echo.Context) pg.DBI {
	db, ok := ctx.Get(databaseContextKey).(pg.DBI)
	if !ok {
		panic("no database on context")
	}

	return db
}

func (c *Controller) getUnauthenticatedRepository(ctx echo.Context) (repository.UnauthenticatedRepository, error) {
	db, ok := ctx.Get(databaseContextKey).(pg.DBI)
	if !ok {
		return nil, errors.Errorf("no transaction for request")
	}

	return repository.NewUnauthenticatedRepository(db), nil
}

func (c *Controller) mustGetUnauthenticatedRepository(ctx echo.Context) repository.UnauthenticatedRepository {
	repo, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		panic(err)
	}

	return repo
}

func (c *Controller) mustGetUserId(ctx echo.Context) uint64 {
	userId := ctx.Values().GetUint64Default(userIdContextKey, 0)
	if userId == 0 {
		panic("unauthorized")
	}

	return userId
}

func (c *Controller) mustGetAccountId(ctx echo.Context) uint64 {
	accountId := ctx.Values().GetUint64Default(accountIdContextKey, 0)
	if accountId == 0 {
		panic("unauthorized")
	}

	return accountId
}

func (c *Controller) getAuthenticatedRepository(ctx echo.Context) (repository.Repository, error) {
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
