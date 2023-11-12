package controller

import (
	"context"
	"strconv"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/captcha"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/monetr/monetr/server/util"
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
	email                    communication.EmailCommunication
	clientTokens             security.ClientTokens
	clock                    clock.Clock
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
	email communication.EmailCommunication,
	clientTokens security.ClientTokens,
	clock clock.Clock,
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
	basicBilling := billing.NewBasicBilling(log, clock, accountsRepo, pubSub)

	plaidWebhookVerification := platypus.NewInMemoryWebhookVerification(log, plaidClient, 5*time.Minute)

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
		email:                    email,
		clientTokens:             clientTokens,
		clock:                    clock,
	}
}

func (c *Controller) Close() error {
	if err := c.plaidWebhookVerification.Close(); err != nil {
		c.log.WithError(err).Error("failed to dispose of plaid webhook verification gracefully")
		return err
	}

	return nil
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
