package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/background"
	"github.com/monetr/monetr/pkg/billing"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/mail"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/stripe_helper"
	"github.com/spf13/cobra"
)

func init() {
	ServeCommand.PersistentFlags().BoolVarP(&MigrateDatabaseFlag, "migrate", "m", false, "Automatically run database migrations on startup. Defaults to: false")
	ServeCommand.PersistentFlags().IntVarP(&PortFlag, "port", "p", 0, "Specify a port to serve HTTP traffic on for monetr.")
	rootCommand.AddCommand(ServeCommand)
}

var (
	PortFlag            int
	MigrateDatabaseFlag = false

	ServeCommand = &cobra.Command{
		Use:   "serve",
		Short: "Run the monetr HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServer()
		},
	}
)

func RunServer() error {
	configuration := config.LoadConfiguration()

	if PortFlag > 0 {
		configuration.Server.ListenPort = PortFlag
	}

	log := logging.NewLoggerWithConfig(configuration.Logging)
	if configFileName := configuration.GetConfigFileName(); configFileName != "" {
		log.WithField("config", configFileName).Info("config file loaded")
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
			Release:          build.Release,
			Environment:      configuration.Environment,
			SampleRate:       configuration.Sentry.SampleRate,
			EnableTracing:    configuration.Sentry.TraceSampleRate > 0,
			TracesSampleRate: configuration.Sentry.TraceSampleRate,
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
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
			scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
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

	db, err := getDatabase(log, configuration, stats)
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

	var stripe stripe_helper.Stripe
	var basicPaywall billing.BasicPayWall
	if configuration.Stripe.Enabled {
		log.Debug("stripe is enabled, creating client")
		stripe = stripe_helper.NewStripeHelperWithCache(
			log,
			configuration.Stripe.APIKey,
			redisCache,
		)

		accountRepo := billing.NewAccountRepository(
			log,
			redisCache,
			db,
		)

		basicPaywall = billing.NewBasicPaywall(log, accountRepo)
	}

	if configuration.Plaid.WebhooksEnabled {
		log.Debugf("plaid webhooks are enabled and will be sent to: %s", configuration.Plaid.WebhooksDomain)
	}

	if configuration.Stripe.Enabled && configuration.Stripe.WebhooksEnabled {
		log.Debugf("stripe webhooks are enabled and will be sent to: %s", configuration.Stripe.WebhooksDomain)
	}

	kms, err := getKMS(log, configuration)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize KMS")
		return err
	}

	plaidSecrets := secrets.NewPostgresPlaidSecretsProvider(log, db, kms)
	plaidClient := platypus.NewPlaid(log, plaidSecrets, repository.NewPlaidRepository(db), configuration.Plaid)

	var smtpClient mail.Communication
	if configuration.Email.Enabled {
		smtpClient = mail.NewSMTPCommunication(log, configuration.Email.SMTP)
	}

	backgroundJobs, err := background.NewBackgroundJobs(
		context.Background(),
		log,
		configuration,
		db,
		redisController.Pool(),
		pubsub.NewPostgresPubSub(log, db),
		plaidClient,
		plaidSecrets,
	)
	if err != nil {
		log.WithError(err).Fatalf("failed to setup background job proceessor")
		return err
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

	app := application.NewApp(configuration, getControllers(
		log,
		configuration,
		db,
		backgroundJobs,
		plaidClient,
		stats,
		stripe,
		redisController.Pool(),
		plaidSecrets,
		basicPaywall,
		smtpClient,
	)...)

	listenAddress := fmt.Sprintf(":%d", configuration.Server.ListenPort)
	go func() {
		if err := app.Start(listenAddress); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to start the server")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	log.Infof("monetr is running, listening at %s", listenAddress)
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
