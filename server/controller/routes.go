package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/monetr/monetr/server/internal/ctxkeys"
	"github.com/monetr/monetr/server/internal/sentryecho"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/util"
	"github.com/sirupsen/logrus"
)

func (c *Controller) RegisterRoutes(app *echo.Echo) {
	if false {
		app.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			OnTimeoutRouteErrorHandler: func(err error, ctx echo.Context) {
				txn, ok := ctx.Get(databaseContextKey).(*pg.Tx)
				if ok {
					log := c.getLog(ctx)
					log.WithError(err).Warn("request timed out, rolling back transaction")
					if terr := txn.Rollback(); terr != nil {
						log.WithError(terr).Error("failed to rollback transaction for timed out request")
					}
				}
			},
			Timeout: 30 * time.Second,
		}))
	}

	api := app.Group(APIPath, middleware.Gzip())

	if c.Stats != nil {
		app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				start := time.Now()
				defer func() {
					c.Stats.FinishedRequest(ctx, time.Since(start))
				}()
				return next(ctx)
			}
		})
	}

	// Generic request logger, log the request being made with a debug level.
	api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if ctx.Path() == APIPath+"/health" {
				return next(ctx)
			}

			// Log after the request completes, this way we have information about the
			// authenticated user (if there is one).
			defer func(ctx echo.Context) {
				log := c.Log.WithFields(logrus.Fields{
					"method":    ctx.Request().Method,
					"path":      ctx.Path(),
					"requestId": util.GetRequestID(ctx),
				})

				claims, err := c.getClaims(ctx)
				if err == nil {
					if claims.LoginId != "" {
						log = log.WithField("loginId", claims.LoginId)
					}
					if claims.AccountId != "" {
						log = log.WithField("accountId", claims.AccountId)
					}
					if claims.UserId != "" {
						log = log.WithField("userId", claims.UserId)
					}
					if claims.Scope != "" {
						log = log.WithField("scope", claims.Scope)
					}
				}

				log.Debugf("%s %s", ctx.Request().Method, ctx.Path())
			}(ctx)

			return next(ctx)
		}
	})

	// Handle both GET and HEAD requests so that uptimerobot doesn't spam error
	// logs.
	api.GET("/health", c.handleHealth)
	api.HEAD("/health", c.handleHealth)

	baseParty := api.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (returnErr error) {
			var span *sentry.Span
			if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
				requestId := util.GetRequestID(ctx)
				hub.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetTag("requestId", requestId)
				})

				tracingCtx := sentry.SetHubOnContext(ctx.Request().Context(), hub)
				name := fmt.Sprintf("%s %s", ctx.Request().Method, ctx.Path())
				span = sentry.StartSpan(
					tracingCtx,
					"http.server",
					sentry.WithTransactionName(name),
					sentry.ContinueFromRequest(ctx.Request()),
				)
				span.Description = name
				span.SetTag("http.method", ctx.Request().Method)
				span.SetTag("http.route", ctx.Path())
				span.SetTag("http.flavor", fmt.Sprintf("%d.%d", ctx.Request().ProtoMajor, ctx.Request().ProtoMinor))
				span.SetTag("http.scheme", ctx.Request().URL.Scheme)
				if userAgent := ctx.Request().UserAgent(); userAgent != "" {
					span.SetTag("http.user_agent", ctx.Request().UserAgent())
				}
				span.SetTag("net.host.name", ctx.Request().URL.Host)

				// Don't sample traces from the icons endpoint right now
				if ctx.Path() == APIPath+"/icons/search" {
					span.Sampled = sentry.SampledFalse
				}

				defer func() {
					if panicErr := recover(); panicErr != nil {
						hub.RecoverWithContext(span.Context(), panicErr)
						c.getLog(ctx).Errorf("panic for request: %+v\n%s", panicErr, string(debug.Stack()))
						returnErr = ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
							"error": "An internal error occurred.",
						})
						span.Status = sentry.SpanStatusInternalError
						// Make sure we always sample error traces
						span.Sampled = sentry.SampledTrue
					} else {
						span.Status = sentry.SpanStatusOK
					}
					span.SetTag("http.status_code", fmt.Sprint(ctx.Response().Status))
					span.Finish()
				}()

				ctx.Set(spanKey, span)

				{
					spanContext := span.Context()
					if requestId != "" { // If there is a request ID, include it on our span context for logging later.
						spanContext = context.WithValue(span.Context(), ctxkeys.RequestID, requestId)
					}
					ctx.Set(spanContextKey, spanContext)
				}

				hub.AddBreadcrumb(&sentry.Breadcrumb{
					Type:     "http",
					Category: ctx.Request().URL.Hostname(),
					Data: map[string]interface{}{
						"url":    ctx.Request().URL.String(),
						"method": ctx.Request().Method,
					},
					Message:   fmt.Sprintf("%s %s", ctx.Request().Method, ctx.Request().URL.String()),
					Level:     "info",
					Timestamp: time.Now(),
				}, nil)
			} else {
				ctx.Set(spanContextKey, ctx.Request().Context())
			}

			log := c.getLog(ctx)
			err := next(ctx)
			if err != nil { // Log the error for the request.
				level := logrus.ErrorLevel
				switch raw := err.(type) {
				case *echo.HTTPError:
					// If this is an error for the user, then don't log at an error level.
					if raw.Code < 500 {
						level = logrus.WarnLevel
					}
				}
				log.WithError(err).Logf(level, "%s", err.Error())
			}

			switch actualError := err.(type) {
			case *json.MarshalerError:
				// TODO, what would happen here, what would be thrown?
			case *echo.HTTPError:
				switch internalError := actualError.Internal.(type) {
				case GenericAPIError:
					if _, ok := internalError.(json.Marshaler); ok {
						return ctx.JSON(actualError.Code, internalError)
					}
				default:
					return ctx.JSON(actualError.Code, map[string]interface{}{
						"error": actualError.Message,
					})
				}
			case nil:
				return err
			default:
				return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error": err.Error(),
				})
			}

			return err
		}
	})

	repoParty := baseParty.Group("", c.databaseRepositoryMiddleware)

	{ // Webhook endpoints
		webhookParty := repoParty.Group("")
		webhookParty.POST("/plaid/webhook", c.postPlaidWebhook)
		webhookParty.POST("/stripe/webhook", c.handleStripeWebhook)
	}

	// unauthed are endpoints that do not require authentication directly, but can
	// still use a token if one is provided.
	unauthed := repoParty.Group("", c.maybeTokenMiddleware)

	// Endpoints used by the client/UI.
	unauthed.GET("/config", c.configEndpoint)

	// These endpoints do not require any authentication.
	unauthed.POST("/authentication/login", c.postLogin)
	unauthed.GET("/authentication/logout", c.logoutEndpoint)
	unauthed.POST("/authentication/register", c.postRegister)
	unauthed.POST("/authentication/verify", c.verifyEndpoint)
	unauthed.POST("/authentication/verify/resend", c.resendVerification)
	unauthed.POST("/authentication/forgot", c.postForgotPassword)
	unauthed.POST("/authentication/reset", c.resetPassword)

	// These endpoints are only accessible if you have a token scoped for MFA.
	multiFactorRequired := repoParty.Group("",
		c.maybeTokenMiddleware,
		c.requireToken(security.MultiFactorScope),
	)
	multiFactorRequired.POST("/authentication/multifactor", c.postMultifactor)

	// You are allowed to request your own user info if you have a token scoped to
	// MFA or just a normally authenticated token.
	userInfo := repoParty.Group("",
		c.maybeTokenMiddleware,
		c.requireToken(security.AuthenticatedScope, security.MultiFactorScope),
	)
	userInfo.GET("/users/me", c.getMe)

	// These endpoints all require a fully authenticated token
	authed := repoParty.Group("",
		c.maybeTokenMiddleware,
		c.requireToken(security.AuthenticatedScope),
	)
	// User
	authed.PUT("/users/security/password", c.changePassword)
	authed.POST("/users/security/totp/setup", c.postSetupTOTP)
	authed.POST("/users/security/totp/confirm", c.postConfirmTOTP)
	// API Keys
	c.RegisterAPIKeyRoutes(authed)
	// Billing
	authed.POST("/billing/create_checkout", c.handlePostCreateCheckout)
	authed.GET("/billing/checkout/:checkoutSessionId", c.handleGetAfterCheckout)
	authed.GET("/billing/portal", c.getBillingPortal)

	billed := authed.Group("", c.requireActiveSubscriptionMiddleware)
	// Icons
	billed.POST("/icons/search", c.searchIcon)
	// Locale and currency data
	billed.GET("/locale/currency", c.listCurrencies)
	// Account
	billed.DELETE("/account", c.deleteAccount)
	// Links
	billed.GET("/links", c.getLinks)
	billed.GET("/links/:linkId", c.getLink)
	billed.POST("/links", c.postLinks)
	billed.PUT("/links/:linkId", c.putLink)
	billed.PUT("/links/convert/:linkId", c.convertLink)
	billed.DELETE("/links/:linkId", c.deleteLink)
	billed.GET("/links/wait/:linkId", c.waitForDeleteLink)
	// Institutions
	billed.GET("/institutions/:institutionId", c.getInstitutionDetails)
	// Bank Accounts
	billed.GET("/bank_accounts", c.getBankAccounts)
	billed.GET("/bank_accounts/:bankAccountId", c.getBankAccount)
	billed.PUT("/bank_accounts/:bankAccountId", c.putBankAccounts)
	billed.GET("/bank_accounts/:bankAccountId/balances", c.getBalances)
	billed.POST("/bank_accounts", c.postBankAccounts)
	// Transactions
	billed.GET("/bank_accounts/:bankAccountId/transactions", c.getTransactions)
	billed.GET("/bank_accounts/:bankAccountId/transactions/:transactionId", c.getTransactionById)
	billed.GET("/bank_accounts/:bankAccountId/transactions/:transactionId/similar", c.getSimilarTransactionsById)
	billed.POST("/bank_accounts/:bankAccountId/transactions", c.postTransactions)
	billed.POST("/bank_accounts/:bankAccountId/transactions/upload", c.postTransactionUpload)
	billed.GET("/bank_accounts/:bankAccountId/transactions/upload/:transactionUploadId", c.getTransactionUploadById)
	billed.GET("/bank_accounts/:bankAccountId/transactions/upload/:transactionUploadId/progress", c.getTransactionUploadProgress)
	billed.PUT("/bank_accounts/:bankAccountId/transactions/:transactionId", c.putTransactions)
	billed.DELETE("/bank_accounts/:bankAccountId/transactions/:transactionId", c.deleteTransactions)
	// Uploads
	billed.GET("/files", c.getFiles)
	// Funding schedules
	billed.GET("/bank_accounts/:bankAccountId/funding_schedules", c.getFundingSchedules)
	billed.GET("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.getFundingScheduleById)
	billed.POST("/bank_accounts/:bankAccountId/funding_schedules", c.postFundingSchedules)
	billed.PUT("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.putFundingSchedules)
	billed.DELETE("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.deleteFundingSchedules)
	// Spending
	billed.GET("/bank_accounts/:bankAccountId/spending", c.getSpending)
	billed.GET("/bank_accounts/:bankAccountId/spending/:spendingId", c.getSpendingById)
	billed.POST("/bank_accounts/:bankAccountId/spending", c.postSpending)
	billed.POST("/bank_accounts/:bankAccountId/spending/transfer", c.postSpendingTransfer)
	billed.PUT("/bank_accounts/:bankAccountId/spending/:spendingId", c.putSpending)
	billed.DELETE("/bank_accounts/:bankAccountId/spending/:spendingId", c.deleteSpending)
	billed.GET("/bank_accounts/:bankAccountId/spending/:spendingId/transactions", c.getSpendingTransactions)
	// Forecasting
	billed.GET("/bank_accounts/:bankAccountId/forecast", c.getForecast)
	billed.POST("/bank_accounts/:bankAccountId/forecast/spending", c.postForecastNewSpending)
	billed.POST("/bank_accounts/:bankAccountId/forecast/next_funding", c.postForecastNextFunding)
	// Plaid Link
	billed.PUT("/plaid/link/update/:linkId", c.putUpdatePlaidLink)
	billed.POST("/plaid/link/update/callback", c.updatePlaidTokenCallback)
	billed.GET("/plaid/link/token/new", c.newPlaidToken)
	billed.POST("/plaid/link/token/callback", c.postPlaidTokenCallback)
	billed.GET("/plaid/link/setup/wait/:linkId", c.getWaitForPlaid)
	billed.POST("/plaid/link/sync", c.postSyncPlaidManually)
}
