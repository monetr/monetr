package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/monetr/monetr/server/internal/ctxkeys"
	"github.com/monetr/monetr/server/internal/sentryecho"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
)

func (c *Controller) RegisterRoutes(app *echo.Echo) {
	if false {
		app.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
			ErrorHandler: func(ctx *echo.Context, err error) error {
				txn, ok := ctx.Get(databaseContextKey).(*pg.Tx)
				if ok {
					log := c.getLog(ctx)
					log.WarnContext(c.getContext(ctx), "request timed out, rolling back transaction", "err", err)
					if terr := txn.Rollback(); terr != nil {
						log.ErrorContext(c.getContext(ctx), "failed to rollback transaction for timed out request", "err", terr)
					}
				}
				return err
			},
			Timeout: 30 * time.Second,
		}))
	}

	api := app.Group(APIPath, middleware.Gzip())

	if c.Stats != nil {
		app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx *echo.Context) error {
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
		return func(ctx *echo.Context) error {
			if ctx.Path() == APIPath+"/health" {
				return next(ctx)
			}

			// Log after the request completes, this way we have information about the
			// authenticated user (if there is one).
			defer func(ctx *echo.Context) {
				log := c.Log.With(
					"method", ctx.Request().Method,
					"path", ctx.Path(),
					"requestId", util.GetRequestID(ctx),
				)

				claims, err := c.getClaims(ctx)
				if err == nil {
					if claims.LoginId != "" {
						log = log.With("loginId", claims.LoginId)
					}
					if claims.AccountId != "" {
						log = log.With("accountId", claims.AccountId)
					}
					if claims.UserId != "" {
						log = log.With("userId", claims.UserId)
					}
					if claims.Scope != "" {
						log = log.With("scope", claims.Scope)
					}
				}

				log.DebugContext(c.getContext(ctx), fmt.Sprintf("%s %s", ctx.Request().Method, ctx.Path()))
			}(ctx)

			return next(ctx)
		}
	})

	// Handle both GET and HEAD requests so that uptimerobot doesn't spam error
	// logs.
	api.GET("/health", c.handleHealth)
	api.HEAD("/health", c.handleHealth)

	baseParty := api.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) (returnErr error) {
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
						c.getLog(ctx).ErrorContext(c.getContext(ctx), fmt.Sprintf("panic for request: %+v\n%s", panicErr, string(debug.Stack())))
						returnErr = ctx.JSON(http.StatusInternalServerError, map[string]any{
							"error": "An internal error occurred.",
						})
						span.Status = sentry.SpanStatusInternalError
						// Make sure we always sample error traces
						span.Sampled = sentry.SampledTrue
					} else {
						span.Status = sentry.SpanStatusOK
					}
					// echo v5 hides the status code behind the http.ResponseWriter,
					// we have to unwrap the *echo.Response to read it.
					status := 0
					if response, err := echo.UnwrapResponse(ctx.Response()); err == nil {
						status = response.Status
					}
					span.SetTag("http.status_code", fmt.Sprint(status))
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
					Data: map[string]any{
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
				level := slog.LevelError
				// If this is a client facing error (status < 500) then don't log
				// at an error level. echo v5 dropped the typed HTTPError we used
				// to assert on here, but both echo's HTTPError and monetr's own
				// apiResponseError expose StatusCode() so we look for that.
				var coder interface{ StatusCode() int }
				if errors.As(err, &coder) && coder.StatusCode() < 500 {
					level = slog.LevelWarn
				}

				// Don't log an error level if we are logging for a context canceled
				// error.
				if errors.Is(err, context.Canceled) {
					level = slog.LevelWarn
				}

				log.Log(c.getContext(ctx), level, err.Error(), "err", err)
			}

			switch err.(type) {
			case nil:
				return nil
			case *json.MarshalerError:
				// TODO, what would happen here, what would be thrown?
				return err
			}

			// monetr's own errors carry a pre shaped body (a validation problems
			// tree, a GenericAPIError, etc) that we want written to the client
			// verbatim. In echo v4 we smuggled this through HTTPError.Message and
			// HTTPError.Internal, but v5 narrowed those so we carry it ourselves.
			var apiErr *apiResponseError
			if errors.As(err, &apiErr) {
				return ctx.JSON(apiErr.code, apiErr.body)
			}

			// echo (and monetr's plain string error helpers) produce HTTPErrors
			// whose message is a string, render those as {"error": message}.
			var httpErr *echo.HTTPError
			if errors.As(err, &httpErr) {
				return ctx.JSON(httpErr.Code, map[string]any{
					"error": httpErr.Message,
				})
			}

			return ctx.JSON(http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
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
	unauthed.POST("/authentication/challenge", c.postChallenge)
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
		c.requireAuthentication(security.MultiFactorScope),
	)
	multiFactorRequired.POST("/authentication/multifactor", c.postMultifactor)

	// You are allowed to request your own user info if you have a token scoped to
	// MFA or just a normally authenticated token.
	userInfo := repoParty.Group("",
		c.maybeApiKeyMiddleware,
		c.maybeTokenMiddleware,
		c.requireAuthentication(security.AuthenticatedScope, security.MultiFactorScope),
	)
	userInfo.GET("/users/me", c.getMe)

	// These endpoints all require a fully authenticated token
	tokenOnly := repoParty.Group("",
		c.maybeTokenMiddleware,
		c.requireAuthentication(security.AuthenticatedScope),
	)
	// User
	tokenOnly.PUT("/users/security/password", c.changePassword)
	tokenOnly.POST("/users/security/totp/setup", c.postSetupTOTP)
	tokenOnly.POST("/users/security/totp/confirm", c.postConfirmTOTP)
	// Billing
	tokenOnly.POST("/billing/create_checkout", c.handlePostCreateCheckout)
	tokenOnly.GET("/billing/checkout/:checkoutSessionId", c.handleGetAfterCheckout)
	tokenOnly.GET("/billing/portal", c.getBillingPortal)

	// Accepts an API Key or a valid session token.
	apiKeyOrToken := repoParty.Group("",
		c.maybeApiKeyMiddleware,
		c.maybeTokenMiddleware,
		c.requireAuthentication(security.AuthenticatedScope),
	)

	billedTokenOnly := tokenOnly.Group("", c.requireActiveSubscriptionMiddleware)
	billedKeyOrToken := apiKeyOrToken.Group("", c.requireActiveSubscriptionMiddleware)

	// API Credentials, reading and deactivating keys can be done without an
	// active subscription. But creating new keys requires an active subscription
	// when billing is enabled.
	tokenOnly.GET("/keys", c.getApiKeys)
	tokenOnly.DELETE("/keys/:apiKeyId", c.deleteApiKey)
	billedTokenOnly.POST("/keys", c.postApiKey)

	// Icons
	billedKeyOrToken.POST("/icons/search", c.searchIcon)
	// Locale and currency data
	billedKeyOrToken.GET("/locale/currency", c.listCurrencies)
	// Account
	billedKeyOrToken.DELETE("/account", c.deleteAccount)
	// Links
	billedKeyOrToken.GET("/links", c.getLinks)
	billedKeyOrToken.GET("/links/:linkId", c.getLink)
	billedKeyOrToken.POST("/links", c.postLinks)
	billedKeyOrToken.PATCH("/links/:linkId", c.patchLink)
	billedKeyOrToken.DELETE("/links/:linkId", c.deleteLink)
	// Institutions, this is not accessible via API keys because it is backed by
	// Plaid and would create a potential spam vector for plaid.
	billedTokenOnly.GET("/institutions/:institutionId", c.getInstitutionDetails)
	// Bank Accounts
	billedKeyOrToken.GET("/bank_accounts", c.getBankAccounts)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId", c.getBankAccount)
	billedKeyOrToken.DELETE("/bank_accounts/:bankAccountId", c.deleteBankAccount)
	billedKeyOrToken.PATCH("/bank_accounts/:bankAccountId", c.patchBankAccount)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/balances", c.getBalances)
	billedKeyOrToken.POST("/bank_accounts", c.postBankAccounts)
	// Transactions
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/transactions", c.getTransactions)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/transactions/:transactionId", c.getTransactionById)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/transactions/:transactionId/similar", c.getSimilarTransactionsById)
	billedKeyOrToken.POST("/bank_accounts/:bankAccountId/transactions", c.postTransactions)
	billedKeyOrToken.POST("/bank_accounts/:bankAccountId/transactions/upload", c.postTransactionUpload)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/transactions/upload/:transactionUploadId", c.getTransactionUploadById)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/transactions/upload/:transactionUploadId/progress", c.getTransactionUploadProgress)
	billedKeyOrToken.PATCH("/bank_accounts/:bankAccountId/transactions/:transactionId", c.patchTransaction)
	billedKeyOrToken.DELETE("/bank_accounts/:bankAccountId/transactions/:transactionId", c.deleteTransactions)

	// Mappings and transaction imports
	if c.Configuration.Features.TransactionImports {
		billedKeyOrToken.GET("/mappings", c.getTransactionImportMappings)
		billedKeyOrToken.POST("/mappings", c.postTransactionImportMapping)
		// Imports
		billedKeyOrToken.POST("/bank_accounts/:bankAccountId/transactions/import", c.postTransactionImport)
		billedKeyOrToken.GET("/bank_accounts/:bankAccountId/transactions/import/:transactionImportId", c.getTransactionImportById)
		billedKeyOrToken.PATCH("/bank_accounts/:bankAccountId/transactions/import/:transactionImportId", c.patchTransactionImport)
	}

	// Uploads
	billedKeyOrToken.GET("/files", c.getFiles)
	// Funding schedules
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/funding_schedules", c.getFundingSchedules)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.getFundingScheduleById)
	billedKeyOrToken.POST("/bank_accounts/:bankAccountId/funding_schedules", c.postFundingSchedules)
	billedKeyOrToken.PATCH("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.patchFundingSchedule)
	billedKeyOrToken.DELETE("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.deleteFundingSchedules)
	// Spending
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/spending", c.getSpending)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/spending/:spendingId", c.getSpendingById)
	billedKeyOrToken.POST("/bank_accounts/:bankAccountId/spending", c.postSpending)
	billedKeyOrToken.POST("/bank_accounts/:bankAccountId/spending/transfer", c.postSpendingTransfer)
	billedKeyOrToken.PUT("/bank_accounts/:bankAccountId/spending/:spendingId", c.putSpending)
	billedKeyOrToken.PATCH("/bank_accounts/:bankAccountId/spending/:spendingId", c.patchSpending)
	billedKeyOrToken.DELETE("/bank_accounts/:bankAccountId/spending/:spendingId", c.deleteSpending)
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/spending/:spendingId/transactions", c.getSpendingTransactions)
	// Forecasting
	billedKeyOrToken.GET("/bank_accounts/:bankAccountId/forecast", c.getForecast)
	billedKeyOrToken.POST("/bank_accounts/:bankAccountId/forecast/spending", c.postForecastNewSpending)
	billedKeyOrToken.POST("/bank_accounts/:bankAccountId/forecast/next_funding", c.postForecastNextFunding)

	// Plaid Link, these endpoints are not accessible via API keys because it
	// creates a potential spam vector for plaid.
	billedTokenOnly.PUT("/plaid/link/update/:linkId", c.putUpdatePlaidLink)
	billedTokenOnly.POST("/plaid/link/update/callback", c.updatePlaidTokenCallback)
	billedTokenOnly.GET("/plaid/link/token/new", c.newPlaidToken)
	billedTokenOnly.POST("/plaid/link/token/callback", c.postPlaidTokenCallback)
	billedTokenOnly.GET("/plaid/link/setup/wait/:linkId", c.getWaitForPlaid)
	billedTokenOnly.POST("/plaid/link/sync", c.postPlaidLinkSync)

	// These endpoints should only be made available when lunch flow is actually
	// enabled in the configuration. This way the endpoints are not available for
	// the hosted version of monetr, but are available for self-hosted instances.
	lunchFlow := billedKeyOrToken.Group("",
		c.requireLunchFlowEnabledMiddleware,
	)
	// Lunch Flow Links
	lunchFlow.GET("/lunch_flow/link", c.getLunchFlowLinks)
	lunchFlow.POST("/lunch_flow/link", c.postLunchFlowLink)
	lunchFlow.GET("/lunch_flow/link/:lunchFlowLinkId", c.getLunchFlowLink)
	lunchFlow.POST("/lunch_flow/link/:lunchFlowLinkId/bank_accounts/refresh", c.postLunchFlowLinkBankAccountsRefresh)
	lunchFlow.GET("/lunch_flow/link/:lunchFlowLinkId/bank_accounts", c.getLunchFlowLinkBankAccounts)
	lunchFlow.POST("/lunch_flow/link/sync", c.postLunchFlowLinkSync)
	lunchFlow.GET("/lunch_flow/link/sync/:linkId/bank_account/:bankAccountId/progress", c.getLunchFlowLinkSyncProgress)
}
