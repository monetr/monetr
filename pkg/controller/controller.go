package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/monetr/monetr/pkg/background"
	"github.com/monetr/monetr/pkg/billing"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/captcha"
	"github.com/monetr/monetr/pkg/communication"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/ctxkeys"
	"github.com/monetr/monetr/pkg/mail"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/stripe_helper"
	"github.com/monetr/monetr/pkg/util"
	"github.com/monetr/monetr/pkg/verification"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	db                       *pg.DB
	configuration            config.Configuration
	captcha                  captcha.Verification
	plaid                    platypus.Platypus
	plaidWebhookVerification platypus.WebhookVerification
	plaidSecrets             secrets.PlaidSecretsProvider
	plaidInstitutions        platypus.PlaidInstitutions
	log                      *logrus.Entry
	jobRunner                background.JobController
	stats                    *metrics.Stats
	stripe                   stripe_helper.Stripe
	ps                       pubsub.PublishSubscribe
	cache                    cache.Cache
	accounts                 billing.AccountRepository
	paywall                  billing.BasicPayWall
	billing                  billing.BasicBilling
	stripeWebhooks           billing.StripeWebhookHandler
	communication            communication.UserCommunication
	emailVerification        verification.Verification
	passwordResetTokens      verification.TokenGenerator
}

func NewController(
	log *logrus.Entry,
	configuration config.Configuration,
	db *pg.DB,
	jobRunner background.JobController,
	plaidClient platypus.Platypus,
	stats *metrics.Stats,
	stripe stripe_helper.Stripe,
	cachePool *redis.Pool,
	plaidSecrets secrets.PlaidSecretsProvider,
	basicPaywall billing.BasicPayWall,
	smtpCommunication mail.Communication,
) *Controller {
	var recaptcha captcha.Verification
	var err error
	if configuration.ReCAPTCHA.Enabled {
		recaptcha, err = captcha.NewReCAPTCHAVerification(
			configuration.ReCAPTCHA.PrivateKey,
		)
		if err != nil {
			panic(err)
		}
	}

	caching := cache.NewCache(log, cachePool)

	accountsRepo := billing.NewAccountRepository(log, caching, db)
	pubSub := pubsub.NewPostgresPubSub(log, db)
	basicBilling := billing.NewBasicBilling(log, accountsRepo, pubSub)

	plaidWebhookVerification := platypus.NewInMemoryWebhookVerification(log, plaidClient, 5*time.Minute)

	var emailVerification verification.Verification
	if configuration.Email.ShouldVerifyEmails() {
		emailVerification = verification.NewEmailVerification(
			log,
			configuration.Email.Verification.TokenLifetime,
			repository.NewEmailRepository(log, db),
			verification.NewTokenGenerator(configuration.Email.Verification.TokenSecret),
		)
	}

	var passwordResetTokenGenerator verification.TokenGenerator
	if configuration.Email.AllowPasswordReset() {
		passwordResetTokenGenerator = verification.NewTokenGenerator(configuration.Email.ForgotPassword.TokenSecret)
	}

	var userCommunication communication.UserCommunication
	if configuration.Email.Enabled {
		userCommunication = communication.NewUserCommunication(
			log,
			configuration,
			smtpCommunication,
		)
	}

	return &Controller{
		captcha:                  recaptcha,
		configuration:            configuration,
		db:                       db,
		plaid:                    plaidClient,
		plaidWebhookVerification: plaidWebhookVerification,
		plaidSecrets:             plaidSecrets,
		plaidInstitutions:        platypus.NewPlaidInstitutionWrapper(log, plaidClient, caching),
		log:                      log,
		jobRunner:                jobRunner,
		stats:                    stats,
		stripe:                   stripe,
		ps:                       pubSub,
		cache:                    caching,
		accounts:                 accountsRepo,
		paywall:                  basicPaywall,
		billing:                  basicBilling,
		stripeWebhooks:           billing.NewStripeWebhookHandler(log, accountsRepo, basicBilling, pubSub),
		communication:            userCommunication,
		emailVerification:        emailVerification,
		passwordResetTokens:      passwordResetTokenGenerator,
	}
}

func (c *Controller) Close() error {
	if err := c.plaidWebhookVerification.Close(); err != nil {
		c.log.WithError(err).Error("failed to dispose of plaid webhook verification gracefully")
		return err
	}

	return nil
}

// @title monetr's REST API
// @version 0.0
// @description This is the REST API for our budgeting application.

// @contact.name Support
// @contact.url http://github.com/monetr/monetr
// @license.name Business Source License 1.1
// @license.url https://github.com/monetr/monetr/blob/main/LICENSE
// @host your.monetr.app/api

// @tag.name Authentication
// @tag.description Authentication endpoints for end users.

// @tag.name Bank Accounts
// @tag.name Billing
// @tag.name Config

// @tag.name Funding Schedules
// @tag.description Funding Schedules are created by the user to tell us when money should be taken from their account to fund their goals and expenses.

// @tag.name Health
// @tag.name Institutions
// @tag.name Links
// @tag.name Plaid
// @tag.name Spending
// @tag.name Transactions

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Cookies
func (c *Controller) RegisterRoutes(app *echo.Echo) {
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

	if c.stats != nil {
		app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				start := time.Now()
				defer func() {
					c.stats.FinishedRequest(ctx, time.Since(start))
				}()
				return next(ctx)
			}
		})
	}

	api := app.Group(APIPath)

	// Generic request logger, log the request being made with a debug level.
	api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if ctx.Path() == APIPath+"/health" {
				return next(ctx)
			}

			log := c.log.WithFields(logrus.Fields{
				"method":    ctx.Request().Method,
				"path":      ctx.Path(),
				"requestId": util.GetRequestID(ctx),
			})

			log.Debugf("%s %s", ctx.Request().Method, ctx.Path())

			return next(ctx)
		}
	})

	// More complete error handler.
	api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
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

				defer func() {
					if panicErr := recover(); panicErr != nil {
						hub.RecoverWithContext(span.Context(), panicErr)
						c.getLog(ctx).Errorf("panic for request: %+v\n%s", panicErr, string(debug.Stack()))
						returnErr = ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
							"error": "An internal error occurred.",
						})
						span.Status = sentry.SpanStatusInternalError
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
					Category: c.configuration.APIDomainName,
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
				log.WithError(err).Errorf("%s", err.Error())
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

	// TODO implement not found error handler.

	api.GET("/NOTICE", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, build.GetNotice())
	})
	api.GET("/health", func(ctx echo.Context) error {
		status := http.StatusOK
		err := c.db.Ping(ctx.Request().Context())
		if err != nil {
			c.getLog(ctx).WithError(err).Warn("failed to ping database")
			status = http.StatusInternalServerError
		}

		result := map[string]interface{}{
			"dbHealthy":  err == nil,
			"apiHealthy": true,
			"revision":   build.Revision,
			"buildTime":  build.BuildTime,
			"serverTime": time.Now().UTC(),
		}

		if build.Release != "" {
			result["release"] = build.Release
		} else {
			result["release"] = nil
		}

		return ctx.JSON(status, result)
	})

	repoParty := api.Group("", c.databaseRepositoryMiddleware)
	// Plaid incoming webhooks
	repoParty.POST("/plaid/webhook", c.handlePlaidWebhook)
	// Stripe incoming webhooks
	repoParty.POST("/stripe/webook", c.handleStripeWebhook)
	repoParty.GET("/sentry", c.getSentryUI)
	repoParty.GET("/config", c.configEndpoint)
	// Authentication
	repoParty.POST("/authentication/login", c.loginEndpoint)
	repoParty.GET("/authentication/logout", c.logoutEndpoint)
	repoParty.POST("/authentication/register", c.registerEndpoint)
	repoParty.POST("/authentication/verify", c.verifyEndpoint)
	repoParty.POST("/authentication/verify/resend", c.resendVerification)
	repoParty.POST("/authentication/forgot", c.sendForgotPassword)
	repoParty.POST("/authentication/reset", c.resetPassword)

	authed := repoParty.Group("", c.authenticationMiddleware)
	// User
	authed.GET("/users/me", c.getMe)
	authed.PUT("/users/security/password", c.changePassword)
	// Billing
	authed.POST("/billing/create_checkout", c.handlePostCreateCheckout)
	authed.GET("/billing/checkout/:checkoutSessionId", c.handleGetAfterCheckout)
	authed.GET("/billing/portal", c.handleGetStripePortal)

	billed := authed.Group("", c.requireActiveSubscriptionMiddleware)
	// Icons
	billed.POST("/icons/search", c.searchIcon)
	// Account
	billed.GET("/account/settings", c.getAccountSettings)
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
	billed.PUT("/bank_accounts/:bankAccountId", c.putBankAccounts)
	billed.GET("/bank_accounts/:bankAccountId/balances", c.getBalances)
	billed.POST("/bank_accounts", c.postBankAccounts)
	// Transactions
	billed.GET("/bank_accounts/:bankAccountId/transactions", c.getTransactions)
	billed.GET("/bank_accounts/:bankAccountId/transactions/spending/:spendingId", c.getTransactionsForSpending)
	billed.POST("/bank_accounts/:bankAccountId/transactions", c.postTransactions)
	billed.PUT("/bank_accounts/:bankAccountId/transactions/:transactionId", c.putTransactions)
	billed.DELETE("/bank_accounts/:bankAccountId/transactions/:transactionId", c.deleteTransactions)
	// Funding schedules
	billed.GET("/bank_accounts/:bankAccountId/funding_schedules", c.getFundingSchedules)
	billed.GET("/bank_accounts/:bankAccountId/funding_schedules/stats", c.getFundingScheduleStats)
	billed.POST("/bank_accounts/:bankAccountId/funding_schedules", c.postFundingSchedules)
	billed.PUT("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.putFundingSchedules)
	billed.DELETE("/bank_accounts/:bankAccountId/funding_schedules/:fundingScheduleId", c.deleteFundingSchedules)
	// Spending
	billed.GET("/bank_accounts/:bankAccountId/spending", c.getSpending)
	billed.POST("/bank_accounts/:bankAccountId/spending", c.postSpending)
	billed.POST("/bank_accounts/:bankAccountId/spending/transfer", c.postSpendingTransfer)
	billed.PUT("/bank_accounts/:bankAccountId/spending/:spendingId", c.putSpending)
	billed.DELETE("/bank_accounts/:bankAccountId/spending/:spendingId", c.deleteSpending)
	// Forecasting
	billed.POST("/bank_accounts/:bankAccountId/forecast/spending", c.postForecastNewSpending)
	billed.POST("/bank_accounts/:bankAccountId/forecast/next_funding", c.postForecastNextFunding)
	// Plaid Link
	billed.PUT("/plaid/link/update/:linkId", c.updatePlaidLink)
	billed.POST("/plaid/link/update/callback", c.updatePlaidTokenCallback)
	billed.GET("/plaid/link/token/new", c.newPlaidToken)
	billed.POST("/plaid/link/token/callback", c.plaidTokenCallback)
	billed.GET("/plaid/link/setup/wait/:linkId", c.waitForPlaid)
	billed.POST("/plaid/link/sync", c.postSyncPlaidManually)
}

func (c *Controller) getContext(ctx echo.Context) context.Context {
	if requestContext, ok := ctx.Get(spanContextKey).(context.Context); ok {
		return requestContext
	}

	return ctx.Request().Context()
}

func (c *Controller) getSpan(ctx echo.Context) *sentry.Span {
	return ctx.Get(spanKey).(*sentry.Span)
}

func (c *Controller) getLog(ctx echo.Context) *logrus.Entry {
	log := c.log.WithContext(c.getContext(ctx)).WithFields(logrus.Fields{
		"requestId": util.GetRequestID(ctx),
	})

	if accountId, ok := ctx.Get(accountIdContextKey).(uint64); ok {
		log = log.WithField("accountId", accountId)
	}

	if userId, ok := ctx.Get(userIdContextKey).(uint64); ok {
		log = log.WithField("userId", userId)
	}

	if loginId, ok := ctx.Get(loginIdContextKey).(uint64); ok {
		log = log.WithField("loginId", loginId)
	}

	return log
}

// reportWrappedError just includes an errors.Wrapf around reportError. It does not modify the response body to include
// the error.
func (c *Controller) reportWrappedError(ctx echo.Context, err error, message string, args ...interface{}) {
	c.reportError(ctx, errors.Wrapf(err, message, args...))
}

// reportError is a simple wrapper to report errors to sentry.io. It is meant to be used to keep track of errors that
// we encounter that we might not want to return to the end user. But might still need in order to diagnose issues.
func (c *Controller) reportError(ctx echo.Context, err error) {
	if spanContext := c.getContext(ctx); spanContext != nil {
		if hub := sentry.GetHubFromContext(spanContext); hub != nil {
			_ = hub.CaptureException(err)
			hub.Scope().SetLevel(sentry.LevelError)
		} else if hub = sentryecho.GetHubFromContext(ctx); hub != nil {
			_ = hub.CaptureException(err)
			hub.Scope().SetLevel(sentry.LevelError)
		} else {
			sentry.CaptureException(err)
		}
	} else {
		if hub := sentryecho.GetHubFromContext(ctx); hub != nil {
			_ = hub.CaptureException(err)
			hub.Scope().SetLevel(sentry.LevelError)
		} else {
			sentry.CaptureException(err)
		}
	}
	c.getLog(ctx).WithError(err).Errorf("error in request: %s", err)
}

func urlParamIntDefault(ctx echo.Context, param string, defaultValue int) int {
	value := ctx.QueryParam(param)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return int(parsed)
}

func urlParamBoolDefault(ctx echo.Context, param string, defaultValue bool) bool {
	value := ctx.QueryParam(param)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
