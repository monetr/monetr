package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/application"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/captcha"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/controller"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/internal/source"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/metrics"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/monetr/monetr/server/ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ServeCommand.PersistentFlags().BoolVarP(&MigrateDatabaseFlag, "migrate", "m", false, "Automatically run database migrations on startup. Defaults to: false")
	ServeCommand.PersistentFlags().BoolVarP(&GenerateCertificates, "generate-certificates", "g", false, "Generate certificates for authentication if they do not already exist. Defaults to: false")
	ServeCommand.PersistentFlags().IntVarP(&PortFlag, "port", "p", 0, "Specify a port to serve HTTP traffic on for monetr.")
	rootCommand.AddCommand(ServeCommand)

	viper.BindPFlag("PostgreSQL.Migrate", ServeCommand.PersistentFlags().Lookup("migrate"))
}

var (
	PortFlag             int
	MigrateDatabaseFlag  = false
	GenerateCertificates = false

	ServeCommand = &cobra.Command{
		Use:   "serve",
		Short: "Run the monetr HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServer()
		},
	}
)

func RunServer() error {
	clock := clock.New()
	configuration := config.LoadConfiguration()

	if PortFlag > 0 {
		configuration.Server.ListenPort = PortFlag
	}

	log := logging.NewLoggerWithConfig(configuration.Logging)
	if configFileName := configuration.GetConfigFileName(); configFileName != "" {
		log.WithField("config", configFileName).Info("config file loaded")
	}

	log.WithFields(logrus.Fields{
		"privateKeyPath":       configuration.Security.PrivateKey,
		"generateCertificates": GenerateCertificates,
	}).Debug("loading certificates")
	publicKey, privateKey, err := loadCertificates(configuration, GenerateCertificates)
	if err != nil {
		log.WithError(err).Fatal("failed to load ed25519 public and private key")
		return err
	}

	clientTokens, err := security.NewPasetoClientTokens(
		log,
		clock,
		configuration.Server.GetBaseURL().String(),
		publicKey,
		privateKey,
	)
	if err != nil {
		log.WithError(err).Fatal("failed to init paseto client tokens interface")
		return err
	}

	if configuration.Plaid.WebhooksEnabled {
		log.WithField("domain", configuration.Plaid.WebhooksDomain).Trace("plaid webhooks are enabled")
	}

	stats := metrics.NewStats()
	stats.Listen(fmt.Sprintf(":%d", configuration.Server.StatsPort))
	defer stats.Close()

	if configuration.Sentry.Enabled {
		log.Debug("sentry is enabled, setting up")
		hostname, err := os.Hostname()
		if err != nil {
			log.WithError(err).Warn("failed to get hostname for sentry")
		}

		err = sentry.Init(sentry.ClientOptions{
			Dsn:              configuration.Sentry.DSN,
			Debug:            false,
			AttachStacktrace: true,
			ServerName:       hostname,
			Dist:             build.Revision,
			Release:          "v" + strings.TrimPrefix(build.Release, "v"),
			Environment:      configuration.Environment,
			SampleRate:       configuration.Sentry.SampleRate,
			EnableTracing:    configuration.Sentry.TraceSampleRate > 0,
			TracesSampleRate: configuration.Sentry.TraceSampleRate,
			Integrations: func(i []sentry.Integration) []sentry.Integration {
				// Add our own contextify frames integration
				return append(i, new(source.ContextifyFramesIntegration))
			},
			BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
				// Make sure user authentication doesn't make its way into sentry.
				if event.Request != nil {
					event.Request.Cookies = ""
					if event.Request.Headers != nil {
						delete(event.Request.Headers, "Authorization")
						delete(event.Request.Headers, "Cookie")
						delete(event.Request.Headers, "Cookies")
						delete(event.Request.Headers, "M-Token")
						delete(event.Request.Headers, "Plaid-Verification")
						delete(event.Request.Headers, "Stripe-Signature")
					}
				}

				return event
			},
		})
		if err != nil {
			log.WithError(err).Error("failed to init sentry")
		}

		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
				if event.Request != nil {
					event.Request.Cookies = ""
					if event.Request.Headers != nil {
						delete(event.Request.Headers, "Authorization")
						delete(event.Request.Headers, "Cookie")
						delete(event.Request.Headers, "Cookies")
						delete(event.Request.Headers, "M-Token")
						delete(event.Request.Headers, "Plaid-Verification")
						delete(event.Request.Headers, "Stripe-Signature")
					}
				}
				return event
			})
		})

		defer sentry.Flush(10 * time.Second)
	}

	db, err := database.GetDatabase(log, configuration, stats)
	if err != nil {
		log.WithError(err).Fatalf("failed to setup database connection: %+v", err)
		return err
	}
	defer db.Close()

	redisController, err := cache.NewRedisCache(log, configuration.Redis)
	if err != nil {
		log.WithError(err).Fatalf("failed to create redis cache: %+v", err)
		return err
	}
	defer redisController.Close()

	redisCache := cache.NewCache(log, redisController.Pool())

	fileStorage, err := setupStorage(log, configuration)
	if err != nil {
		log.WithError(err).Fatal("could not setup file storage")
		return err
	}

	var pubSub pubsub.PublishSubscribe
	{ // At the moment only postgresql publish and subscribe is supported.
		pubSub = pubsub.NewPostgresPubSub(log, db)
	}

	var accountsRepo repository.AccountsRepository
	{ // Create the accounts repository that will be used by many things.
		accountsRepo = repository.NewAccountRepository(
			log,
			redisCache,
			db,
		)
	}

	var stripe stripe_helper.Stripe
	var bill billing.Billing
	if configuration.Stripe.IsBillingEnabled() {
		log.Debug("stripe is enabled, creating client")
		stripe = stripe_helper.NewStripeHelperWithCache(
			log,
			configuration.Stripe.APIKey,
			redisCache,
		)

		bill = billing.NewBilling(
			log,
			clock,
			configuration,
			accountsRepo,
			stripe,
			pubSub,
		)
	}

	kms, err := getKMS(log, configuration)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize KMS")
		return err
	}

	var plaidClient *platypus.Plaid
	if configuration.Plaid.Enabled {
		log.Debug("plaid is enabled and will be setup")
		if configuration.Plaid.WebhooksEnabled {
			log.Debugf("plaid webhooks are enabled and will be sent to: %s", configuration.Plaid.WebhooksDomain)
		}
		plaidClient = platypus.NewPlaid(log, clock, kms, db, configuration.Plaid)
	}

	plaidInstitutions := platypus.NewPlaidInstitutionWrapper(
		log,
		plaidClient,
		redisCache,
	)

	plaidWebhooks := platypus.NewInMemoryWebhookVerification(
		log,
		plaidClient,
		1*time.Hour,
	)

	var recaptcha captcha.Verification
	if configuration.ReCAPTCHA.Enabled {
		recaptcha, err = captcha.NewReCAPTCHAVerification(
			configuration.ReCAPTCHA.PrivateKey,
		)
		if err != nil {
			panic(err)
		}
	}

	var email communication.EmailCommunication
	if configuration.Email.Enabled {
		email = communication.NewEmailCommunication(log, configuration)
	}

	var backgroundJobs *background.BackgroundJobs
	{ // Setup the background job processor with a 30 second timeout.
		withTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		backgroundJobs, err = background.NewBackgroundJobs(
			withTimeout,
			log,
			clock,
			configuration,
			db,
			pubSub,
			plaidClient,
			kms,
			fileStorage,
			bill,
			email,
		)
		if err != nil {
			cancel()
			log.WithError(err).Fatalf("failed to setup background job proceessor")
			return err
		}
		cancel()
	}

	if err = backgroundJobs.Start(); err != nil {
		log.WithError(err).Fatalf("failed to start background job worker")
		return err
	}
	defer func() {
		if err := backgroundJobs.Close(); err != nil {
			log.WithError(err).Error("failed to close background jobs processor gracefully")
		}
	}()

	applicationControllers := []application.Controller{
		&controller.Controller{
			Accounts:                 accountsRepo,
			Billing:                  bill,
			Cache:                    redisCache,
			Captcha:                  recaptcha,
			ClientTokens:             clientTokens,
			Clock:                    clock,
			Configuration:            configuration,
			DB:                       db,
			Email:                    email,
			FileStorage:              fileStorage,
			JobRunner:                backgroundJobs,
			KMS:                      kms,
			Log:                      log,
			Plaid:                    plaidClient,
			PlaidInstitutions:        plaidInstitutions,
			PlaidWebhookVerification: plaidWebhooks,
			PubSub:                   pubSub,
			Stats:                    stats,
			Stripe:                   stripe,
		},
		ui.NewUIController(log, configuration),
	}

	// Create the actual application for echo to run.
	app := application.NewApp(configuration, applicationControllers...)

	// Register the API key middleware
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	RegisterAPIKeyMiddleware(app, apiKeyRepo, db)

	protocol := "http"
	if configuration.Server.TLSCertificate != "" && configuration.Server.TLSKey != "" {
		protocol = "https"
	}
	listenAddress := fmt.Sprintf("%s:%d", configuration.Server.ListenAddress, configuration.Server.ListenPort)
	go func() {
		var err error
		if configuration.Server.TLSCertificate != "" && configuration.Server.TLSKey != "" {
			log.Info("server will start a TLS listener")
			err = app.StartTLS(
				listenAddress,
				configuration.Server.TLSCertificate,
				configuration.Server.TLSKey,
			)
		} else {
			err = app.Start(listenAddress)
		}
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to start the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	log.WithFields(logrus.Fields{
		"listenAddress":   fmt.Sprintf("%s://%s", protocol, listenAddress),
		"externalAddress": configuration.Server.GetBaseURL().String(),
	}).Info("monetr is running")

	<-quit
	log.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("failed to gracefully shutdown the server")
	}
	log.Info("http server shutdown complete")

	return nil
}
