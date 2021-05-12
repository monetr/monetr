package controller

import (
	"context"
	"github.com/getsentry/sentry-go"
	sentryiris "github.com/getsentry/sentry-go/iris"
	"github.com/gomodule/redigo/redis"
	"github.com/monetrapp/rest-api/pkg/build"
	"github.com/monetrapp/rest-api/pkg/internal/plaid_helper"
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/monetrapp/rest-api/pkg/pubsub"
	stripe_client "github.com/stripe/stripe-go/v72/client"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/xlzd/gotp"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

const (
	TokenName = "M-Token"
)

type Controller struct {
	db             *pg.DB
	configuration  config.Configuration
	captcha        *recaptcha.ReCAPTCHA
	plaid          plaid_helper.Client
	smtp           *smtp.Client
	mailVerifyCode *gotp.HOTP
	log            *logrus.Entry
	job            jobs.JobManager
	stats          *metrics.Stats
	stripeClient   *stripe_client.API
	ps             pubsub.PublishSubscribe
	cache          *redis.Pool
}

func NewController(
	log *logrus.Entry,
	configuration config.Configuration,
	db *pg.DB,
	job jobs.JobManager,
	plaidClient plaid_helper.Client,
	stats *metrics.Stats,
	stripeClient *stripe_client.API,
	cache *redis.Pool,
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

	return &Controller{
		captcha:       &captcha,
		configuration: configuration,
		db:            db,
		plaid:         plaidClient,
		log:           log,
		job:           job,
		stats:         stats,
		stripeClient:  stripeClient,
		ps:            pubsub.NewPostgresPubSub(log, db),
		cache:         cache,
	}
}

// @title monetr's REST API
// @version 0.0
// @description This is the REST API for our budgeting application.

// @contact.name Support
// @contact.url http://github.com/monetrapp/rest-api

// @license.name Business Source License 1.1
// @license.url https://github.com/monetrapp/rest-api/blob/main/LICENSE

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

	app.Get("/health", c.getHealth)

	app.PartyFunc(APIPath, func(p router.Party) {
		p.Use(c.loggingMiddleware)
		p.OnAnyErrorCode(func(ctx iris.Context) {
			if err := ctx.GetErr(); err != nil {
				if hub := sentryiris.GetHubFromContext(ctx); hub != nil {
					_ = hub.CaptureException(err)
				} else {
					sentry.CaptureException(err)
				}

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
				ctx.JSON(map[string]interface{}{
					"error": err.Error(),
				})
			}
		})

		// Trace API calls to sentry
		p.Use(func(ctx iris.Context) {
			if hub := sentryiris.GetHubFromContext(ctx); hub != nil {
				tracingCtx := sentry.SetHubOnContext(ctx.Request().Context(), hub)
				name := strings.TrimSpace(strings.TrimPrefix(ctx.RouteName(), ctx.Method()))
				span := sentry.StartSpan(tracingCtx, ctx.Method(), sentry.TransactionName(name))
				defer span.Finish()

				ctx.Values().Set(spanContextKey, span.Context())
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

			repoParty.PartyFunc("/authentication", func(repoParty router.Party) {
				repoParty.Post("/login", c.loginEndpoint)
				repoParty.Post("/register", c.registerEndpoint)
				repoParty.Post("/verify", c.verifyEndpoint)
			})

			repoParty.Use(c.authenticationMiddleware)

			repoParty.PartyFunc("/users", c.handleUsers)
			repoParty.PartyFunc("/links", c.linksController)
			repoParty.PartyFunc("/plaid/link", c.handlePlaidLinkEndpoints)

			repoParty.PartyFunc("/bank_accounts", func(bankParty router.Party) {
				c.handleBankAccounts(bankParty)
				c.handleTransactions(bankParty)
				c.handleFundingSchedules(bankParty)
				c.handleSpending(bankParty)
			})

			repoParty.PartyFunc("/jobs", c.handleJobs)

			if c.configuration.Environment != "production" {
				repoParty.Get("/test/error", func(ctx iris.Context) {
					c.badRequest(ctx, "this endpoint is meant to be used to test error reporting to sentry")
				})
			}

			if c.configuration.Stripe.Enabled {
				repoParty.PartyFunc("/billing", c.handleBilling)
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
