package controller

import (
	"context"
	"strconv"

	"fmt"
	"log/slog"

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
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/storage"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/monetr/monetr/server/util"

	"github.com/pkg/errors"
)

type Controller struct {
	Accounts      repository.AccountsRepository
	Billing       billing.Billing
	Cache         cache.Cache
	Captcha       captcha.Verification
	ClientTokens  security.ClientTokens
	Clock         clock.Clock
	Configuration config.Configuration
	DB            *pg.DB
	Email         communication.EmailCommunication
	FileStorage   storage.Storage
	// Deprecated: Use [Controller.Queue] instead!
	JobRunner                background.JobController
	Queue                    queue.Enqueuer
	KMS                      secrets.KeyManagement
	Log                      *slog.Logger
	Plaid                    platypus.Platypus
	PlaidInstitutions        platypus.PlaidInstitutions
	PlaidWebhookVerification platypus.WebhookVerification
	PubSub                   pubsub.PublishSubscribe
	Stats                    *metrics.Stats
	Stripe                   stripe_helper.Stripe
}

func (c *Controller) Close() error {
	if err := c.PlaidWebhookVerification.Close(); err != nil {
		c.Log.ErrorContext(context.Background(), "failed to dispose of plaid webhook verification gracefully", "err", err)
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

func (c *Controller) getLog(ctx echo.Context) *slog.Logger {
	log := c.Log.With("requestId", util.GetRequestID(ctx))

	claims, ok := ctx.Get(authenticationKey).(security.Claims)
	if !ok {
		return log
	}

	if claims.AccountId != "" {
		log = log.With("accountId", claims.AccountId)
	}

	if claims.UserId != "" {
		log = log.With("userId", claims.UserId)
	}

	if claims.LoginId != "" {
		log = log.With("loginId", claims.LoginId)
	}

	return log
}

// reportWrappedError just includes an errors.Wrapf around reportError. It does not modify the response body to include
// the error.
func (c *Controller) reportWrappedError(ctx echo.Context, err error, message string, args ...any) {
	c.reportError(ctx, errors.Wrapf(err, message, args...))
}

// reportError is a simple wrapper to report errors to sentry.io. It is meant to
// be used to keep track of errors that we encounter that we might not want to
// return to the end user. But might still need in order to diagnose issues.
func (c *Controller) reportError(ctx echo.Context, err error) {
	level := sentry.LevelError
	report := true
	if errors.Is(err, context.Canceled) {
		level = sentry.LevelWarning
		report = false
	}
	var hub *sentry.Hub
	// Try to derive the hub from the current span's context if possible.
	if spanContext := c.getContext(ctx); spanContext != nil {
		hub = sentry.GetHubFromContext(spanContext)
	}
	// But if we can't then try to derive the hub from the normal context.
	if hub == nil {
		hub = sentryecho.GetHubFromContext(ctx)
	}

	// If hub is defined then capture the exception here.
	if hub != nil {
		_ = hub.CaptureException(err)
		// Use the level from above, if we are dealing with a dumb error that isn't
		// important we don't want to set an error level.
		hub.Scope().SetLevel(level)
	} else if report {
		sentry.CaptureException(err)
	}

	reqCtx := c.getContext(ctx)
	switch level {
	case sentry.LevelError:
		c.getLog(ctx).ErrorContext(reqCtx, fmt.Sprintf("error in request: %s", err), "err", err)
	case sentry.LevelWarning:
		c.getLog(ctx).WarnContext(reqCtx, fmt.Sprintf("error in request: %s", err), "err", err)
	default:
		c.getLog(ctx).DebugContext(reqCtx, fmt.Sprintf("error in request: %s", err), "err", err)
	}
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
