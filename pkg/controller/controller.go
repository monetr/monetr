package controller

import (
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"net/http"
	"net/smtp"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"github.com/xlzd/gotp"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

const (
	TokenName = "H-Token"
)

type Controller struct {
	db             *pg.DB
	configuration  config.Configuration
	captcha        *recaptcha.ReCAPTCHA
	plaid          *plaid.Client
	smtp           *smtp.Client
	mailVerifyCode *gotp.HOTP
	log            *logrus.Entry
	job            jobs.JobManager
	stats          *metrics.Stats
}

func NewController(
	configuration config.Configuration,
	db *pg.DB,
	job jobs.JobManager,
	plaidClient *plaid.Client,
	stats *metrics.Stats,
) *Controller {
	logger := logrus.New()
	level, err := logrus.ParseLevel(configuration.Logging.Level)
	if err != nil {
		panic(err)
	}
	logger.SetLevel(level)
	entry := logrus.NewEntry(logger)
	var captcha recaptcha.ReCAPTCHA
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
		log:           entry,
		job:           job,
		stats:         stats,
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
// @name H-Token
func (c *Controller) RegisterRoutes(app *iris.Application) {
	if c.stats != nil {
		app.UseGlobal(func(ctx *context.Context) {
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
		p.OnAnyErrorCode(func(ctx *context.Context) {
			if err := ctx.GetErr(); err != nil {
				ctx.JSON(map[string]interface{}{
					"error": err.Error(),
				})
			}
		})
		p.OnErrorCode(http.StatusNotFound, func(ctx *context.Context) {
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

		if c.configuration.EnableWebhooks {
			// Webhooks use their own authentication, so we want to declare this first.
			p.Post("/plaid/webhook/{identifier:string}", c.handlePlaidWebhook)
		}

		// For the following endpoints we want to have a repository available to us.
		p.PartyFunc("/", func(repoParty router.Party) {
			repoParty.Use(c.setupRepositoryMiddleware)
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
func (c *Controller) getHealth(ctx *context.Context) {
	err := c.db.Ping(ctx.Request().Context())

	ctx.JSON(map[string]interface{}{
		"dbHealthy":  err == nil,
		"apiHealthy": true,
	})
}
