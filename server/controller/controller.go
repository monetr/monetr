package controller

import (
	"context"
	"strconv"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/captcha"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/sentryecho"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/storage"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	Accounts                 repository.AccountsRepository
	Billing                  billing.Billing
	Cache                    cache.Cache
	Captcha                  captcha.Verification
	ClientTokens             security.ClientTokens
	Clock                    clock.Clock
	Configuration            config.Configuration
	DB                       *pg.DB
	Email                    communication.EmailCommunication
	FileStorage              storage.Storage
	JobRunner                background.JobController
	KMS                      secrets.KeyManagement
	Log                      *logrus.Entry
	Plaid                    platypus.Platypus
	PlaidInstitutions        platypus.PlaidInstitutions
	PlaidWebhookVerification platypus.WebhookVerification
	PubSub                   pubsub.PublishSubscribe
	Stats                    *metrics.Stats
	Stripe                   stripe_helper.Stripe
}

func (c *Controller) Close() error {
	if err := c.PlaidWebhookVerification.Close(); err != nil {
		c.Log.WithError(err).Error("failed to dispose of plaid webhook verification gracefully")
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
	log := c.Log.WithContext(c.getContext(ctx)).WithFields(logrus.Fields{
		"requestId": util.GetRequestID(ctx),
	})

	claims, ok := ctx.Get(authenticationKey).(security.Claims)
	if !ok {
		return log
	}

	if claims.AccountId != "" {
		log = log.WithField("accountId", claims.AccountId)
	}

	if claims.UserId != "" {
		log = log.WithField("userId", claims.UserId)
	}

	if claims.LoginId != "" {
		log = log.WithField("loginId", claims.LoginId)
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
	parsed, err := strconv.ParseInt(value, 10, 32)
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
