package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetr/monetr/pkg/billing"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/communication"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/platypus"
	"github.com/monetr/monetr/pkg/internal/stripe_helper"
	"github.com/monetr/monetr/pkg/jobs"
	"github.com/monetr/monetr/pkg/mail"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/verification"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xlzd/gotp"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

const (
	TokenName = "M-Token"
)

type Controller struct {
	db                       *pg.DB
	configuration            config.Configuration
	captcha                  *recaptcha.ReCAPTCHA
	plaid                    platypus.Platypus
	plaidWebhookVerification platypus.WebhookVerification
	plaidSecrets             secrets.PlaidSecretsProvider
	smtp                     *smtp.Client
	mailVerifyCode           *gotp.HOTP
	log                      *logrus.Entry
	job                      jobs.JobManager
	stats                    *metrics.Stats
	stripe                   stripe_helper.Stripe
	ps                       pubsub.PublishSubscribe
	cache                    *redis.Pool
	accounts                 billing.AccountRepository
	paywall                  billing.BasicPayWall
	billing                  billing.BasicBilling
	stripeWebhooks           billing.StripeWebhookHandler
	communication            communication.UserCommunication
	emailVerification        verification.Verification
}

func NewController(
	log *logrus.Entry,
	configuration config.Configuration,
	db *pg.DB,
	job jobs.JobManager,
	plaidClient platypus.Platypus,
	stats *metrics.Stats,
	stripe stripe_helper.Stripe,
	cachePool *redis.Pool,
	plaidSecrets secrets.PlaidSecretsProvider,
	basicPaywall billing.BasicPayWall,
	smtpCommunication mail.Communication,
) *Controller {
	var captcha recaptcha.ReCAPTCHA
	var err error
	if configuration.ReCAPTCHA.Enabled {
		captcha, err = recaptcha.NewReCAPTCHA(
			configuration.ReCAPTCHA.PrivateKey,
			recaptcha.V2,
			30*time.Second,
		)
		if err != nil {
			panic(err)
		}
	}

	accountsRepo := billing.NewAccountRepository(log, cache.NewCache(log, cachePool), db)
	pubSub := pubsub.NewPostgresPubSub(log, db)
	basicBilling := billing.NewBasicBilling(log, accountsRepo, pubSub)

	plaidWebhookVerification := platypus.NewInMemoryWebhookVerification(log, plaidClient, 5*time.Minute)

	var emailVerification verification.Verification
	if configuration.Email.ShouldVerifyEmails() {
		emailVerification = verification.NewEmailVerification(
			log,
			configuration.Email.Verification.TokenLifetime,
			repository.NewEmailRepository(log, db),
			verification.NewJWTEmailVerification(configuration.Email.Verification.TokenSecret),
		)
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
		captcha:                  &captcha,
		configuration:            configuration,
		db:                       db,
		plaid:                    plaidClient,
		plaidWebhookVerification: plaidWebhookVerification,
		plaidSecrets:             plaidSecrets,
		log:                      log,
		job:                      job,
		stats:                    stats,
		stripe:                   stripe,
		ps:                       pubSub,
		cache:                    cachePool,
		accounts:                 accountsRepo,
		paywall:                  basicPaywall,
		billing:                  basicBilling,
		stripeWebhooks:           billing.NewStripeWebhookHandler(log, accountsRepo, basicBilling, pubSub),
		communication:            userCommunication,
		emailVerification:        emailVerification,
	}
}

// @title monetr's REST API
// @version 0.0
// @description This is the REST API for our budgeting application.

// @contact.name Support
// @contact.url http://github.com/monetr/monetr
// @license.name Business Source License 1.1
// @license.url https://github.com/monetr/monetr/blob/main/LICENSE
// @host api.monetr.app

// @tag.name Funding Schedules
// @tag.description Funding Schedules are created by the user to tell us when money should be taken from their account to fund their goals and expenses.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name M-Token
func (c *Controller) RegisterRoutes(app *iris.Application) {
	if c.stats != nil {
		app.UseGlobal(func(ctx iris.Context) {
			start := time.Now()
			defer func() {
				c.stats.FinishedRequest(ctx, time.Since(start))
			}()

			ctx.Next()
		})
	}

	app.Use(func(ctx iris.Context) {
		if ctx.Path() == APIPath + "/health" {
			ctx.Next()
			return
		}

		log := c.log.WithFields(logrus.Fields{
			"requestId": ctx.GetHeader("X-Request-Id"),
		})

		log.Debug(ctx.RouteName())

		ctx.Next()
	})

	app.PartyFunc(APIPath, func(p router.Party) {
		p.Get("/health", c.getHealth)

		p.Use(c.loggingMiddleware)
		p.OnAnyErrorCode(func(ctx iris.Context) {
			if err := ctx.GetErr(); err != nil {
				c.reportError(ctx, err)

				ctx.JSON(map[string]interface{}{
					"error": err.Error(),
				})
			}
		})
		p.OnErrorCode(http.StatusNotFound, func(ctx iris.Context) {
			if err := ctx.GetErr(); err == nil {
				ctx.JSON(map[string]interface{}{
					"path":  ctx.Path(),
					"error": "the requested path does not exist",
				})
			} else {
				c.reportError(ctx, err)
				ctx.JSON(map[string]interface{}{
					"error": err.Error(),
				})
			}
		})

		// Trace API calls to sentry
		p.Use(func(ctx iris.Context) {
			var span *sentry.Span
			if hub := sentryiris.GetHubFromContext(ctx); hub != nil {
				if requestId := ctx.GetHeader("X-Request-Id"); requestId != "" {
					hub.ConfigureScope(func(scope *sentry.Scope) {
						scope.SetTag("requestId", requestId)
					})
				}

				tracingCtx := sentry.SetHubOnContext(ctx.Request().Context(), hub)
				name := strings.TrimSpace(strings.TrimPrefix(ctx.RouteName(), ctx.Method()))
				span = sentry.StartSpan(
					tracingCtx,
					ctx.Method(),
					sentry.TransactionName(name),
					sentry.ContinueFromRequest(ctx.Request()),
				)
				span.Description = strings.TrimSpace(strings.TrimPrefix(ctx.RouteName(), ctx.Method()))
				defer func() {
					if span.Status == sentry.SpanStatusUndefined {
						if ctx.GetErr() != nil {
							span.Status = sentry.SpanStatusInternalError
						} else {
							span.Status = sentry.SpanStatusOK
						}
					}
					span.Finish()
				}()

				ctx.Values().Set(spanKey, span)
				ctx.Values().Set(spanContextKey, span.Context())

				hub.AddBreadcrumb(&sentry.Breadcrumb{
					Type:     "http",
					Category: c.configuration.APIDomainName,
					Data: map[string]interface{}{
						"url":    ctx.Request().URL.String(),
						"method": ctx.Method(),
					},
					Message:   fmt.Sprintf("%s %s", ctx.Method(), ctx.Request().URL.String()),
					Level:     "info",
					Timestamp: time.Now(),
				}, nil)
			} else {
				ctx.Values().Set(spanContextKey, ctx.Request().Context())
			}

			ctx.Next()
		})

		// For the following endpoints we want to have a repository available to us.
		p.PartyFunc("/", func(repoParty router.Party) {
			repoParty.Use(c.setupRepositoryMiddleware)

			if c.configuration.Plaid.WebhooksEnabled {
				// Webhooks use their own authentication, so we want to declare this first.
				repoParty.Post("/plaid/webhook", c.handlePlaidWebhook)
			}

			if c.configuration.Stripe.Enabled && c.configuration.Stripe.WebhooksEnabled {
				repoParty.PartyFunc("/stripe", c.handleStripe)
			}

			repoParty.Get("/config", c.configEndpoint)

			repoParty.PartyFunc("/authentication", c.handleAuthentication)

			repoParty.Use(c.authenticationMiddleware)

			repoParty.PartyFunc("/users", c.handleUsers)
			if c.configuration.Stripe.Enabled {
				repoParty.PartyFunc("/billing", c.handleBilling)

				// All endpoints after this require verification that the user has an active subscription.
				if c.configuration.Stripe.IsBillingEnabled() {
					repoParty.Use(c.requireActiveSubscriptionMiddleware)
				}
			}

			repoParty.PartyFunc("/links", c.linksController)
			repoParty.PartyFunc("/bank_accounts", func(bankParty router.Party) {
				c.handleBankAccounts(bankParty)
				c.handleTransactions(bankParty)
				c.handleFundingSchedules(bankParty)
				c.handleSpending(bankParty)
			})

			repoParty.PartyFunc("/plaid/link", c.handlePlaidLinkEndpoints)

			if c.configuration.Environment != "production" {
				repoParty.Get("/test/error", func(ctx iris.Context) {
					c.badRequest(ctx, "this endpoint is meant to be used to test error reporting to sentry")
				})
			}
		})
	})

}

// Check API Health
// @Summary Check API Health
// @ID api-health
// @tags Health
// @description Just a simple health check endpoint. This is not used at all in the frontend of the application and is meant to be used in containers to determine if the primary API listener is working.
// @Produce json
// @Router /health [get]
// @Success 200 {object} swag.HealthResponse
func (c *Controller) getHealth(ctx iris.Context) {
	err := c.db.Ping(ctx.Request().Context())
	if err != nil {
		c.getLog(ctx).WithError(err).Warn("failed to ping database")
	}

	result := map[string]interface{}{
		"dbHealthy":  err == nil,
		"apiHealthy": true,
		"revision":   build.Revision,
		"buildTime":  build.BuildTime,
	}

	if build.Release != "" {
		result["release"] = build.Release
	} else {
		result["release"] = nil
	}

	ctx.JSON(result)
}

func (c *Controller) getContext(ctx iris.Context) context.Context {
	return ctx.Values().Get(spanContextKey).(context.Context)
}

func (c *Controller) getSpan(ctx iris.Context) *sentry.Span {
	return ctx.Values().Get(spanKey).(*sentry.Span)
}

func (c *Controller) getLog(ctx iris.Context) *logrus.Entry {
	log := c.log.WithFields(logrus.Fields{
		"requestId": ctx.GetHeader("X-Request-Id"),
	})

	if accountId := ctx.Values().GetUint64Default(accountIdContextKey, 0); accountId > 0 {
		log = log.WithField("accountId", accountId)
	}

	if userId := ctx.Values().GetUint64Default(userIdContextKey, 0); userId > 0 {
		log = log.WithField("userId", userId)
	}

	if loginId := ctx.Values().GetUint64Default(loginIdContextKey, 0); loginId > 0 {
		log = log.WithField("loginId", loginId)
	}

	return log
}

// reportWrappedError just includes an errors.Wrapf around reportError.
func (c *Controller) reportWrappedError(ctx iris.Context, err error, message string, args ...interface{}) {
	c.reportError(ctx, errors.Wrapf(err, message, args...))
}

// reportError is a simple wrapper to report errors to sentry.io. It is meant to be used to keep track of errors that
// we encounter that we might not want to return to the end user. But might still need in order to diagnose issues.
func (c *Controller) reportError(ctx iris.Context, err error) {
	if spanContext := c.getContext(ctx); spanContext != nil {
		if hub := sentry.GetHubFromContext(spanContext); hub != nil {
			_ = hub.CaptureException(err)
			hub.Scope().SetLevel(sentry.LevelError)
		} else {
			sentry.CaptureException(err)
		}
	} else {
		if hub := sentryiris.GetHubFromContext(ctx); hub != nil {
			_ = hub.CaptureException(err)
			hub.Scope().SetLevel(sentry.LevelError)
		} else {
			sentry.CaptureException(err)
		}
	}
}
