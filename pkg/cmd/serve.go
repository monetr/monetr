package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
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
	"github.com/monetr/monetr/pkg/vault_helper"
	"github.com/spf13/cobra"
)

func init() {
	ServeCommand.PersistentFlags().BoolVarP(&MigrateDatabaseFlag, "migrate", "m", false, "Automatically run database migrations on startup. Defaults to: false")
	ServeCommand.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Specify a config file to use, if omitted ./config.yaml or /etc/monetr/config.yaml will be used.")
	ServeCommand.PersistentFlags().IntVarP(&PortFlag, "port", "p", 0, "Specify a port to serve HTTP traffic on for monetr.")
	rootCommand.AddCommand(ServeCommand)
}

var (
	PortFlag            int
	configFilePath      = ""
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
	var configPath *string
	if len(configFilePath) > 0 {
		configPath = &configFilePath
	}

	configuration := config.LoadConfiguration(configPath)

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

	var vault vault_helper.VaultHelper
	if configuration.Vault.Enabled {
		log.Debug("vault is enabled for secret storage")
		client, err := vault_helper.NewVaultHelper(log, vault_helper.Config{
			Address:         configuration.Vault.Address,
			Role:            configuration.Vault.Role,
			Auth:            configuration.Vault.Auth,
			Token:           configuration.Vault.Token,
			TokenFile:       configuration.Vault.TokenFile,
			Timeout:         configuration.Vault.Timeout,
			IdleConnTimeout: configuration.Vault.IdleConnTimeout,
			Username:        configuration.Vault.Username,
			Password:        configuration.Vault.Password,
		})
		if err != nil {
			log.WithError(err).Fatalf("failed to create vault helper")
			return err
		}

		vault = client
	}

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

	var plaidSecrets secrets.PlaidSecretsProvider
	if configuration.Vault.Enabled {
		log.Debugf("secrets will be stored in vault")
		plaidSecrets = secrets.NewVaultPlaidSecretsProvider(log, vault)
	} else {
		log.Debugf("secrets will be stored in postgres")
		plaidSecrets = secrets.NewPostgresPlaidSecretsProvider(log, db)
	}

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

	idleConnsClosed := make(chan struct{})
	iris.RegisterOnInterrupt(func() {
		log.Info("shutting down")
		timeout := 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// close all hosts.
		_ = app.Shutdown(ctx)
		log.Info("http server shutdown complete")
		close(idleConnsClosed)
	})

	listenAddress := fmt.Sprintf(":%d", configuration.Server.ListenPort)
	_ = app.Listen(listenAddress, iris.WithoutInterruptHandler, iris.WithoutServerError(iris.ErrServerClosed))

	<-idleConnsClosed

	return nil
}
