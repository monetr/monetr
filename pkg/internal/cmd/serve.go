package cmd

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/billing"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/certhelper"
	"github.com/monetr/monetr/pkg/internal/migrations"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/internal/platypus"
	"github.com/monetr/monetr/pkg/internal/stripe_helper"
	"github.com/monetr/monetr/pkg/internal/vault_helper"
	"github.com/monetr/monetr/pkg/jobs"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/monetr/monetr/pkg/mail"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	ServeCommand.PersistentFlags().BoolVarP(&migrateDatabase, "migrate", "m", false, "Automatically run database migrations on startup. Defaults to: false")
	ServeCommand.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Specify a config file to use, if omitted ./config.yaml or /etc/monetr/config.yaml will be used.")
	RootCommand.AddCommand(ServeCommand)
}

var (
	configFilePath  = ""
	migrateDatabase = false

	ServeCommand = &cobra.Command{
		Use:   "serve",
		Short: "Run the REST API HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServer()
		},
	}
)

func RunServer() error {
	stats := metrics.NewStats()
	// TODO Allow stats port to be configurable.
	stats.Listen(":9000")
	defer stats.Close()

	var configPath *string
	if len(configFilePath) > 0 {
		configPath = &configFilePath
	}

	configuration := config.LoadConfiguration(configPath)

	log := logging.NewLoggerWithLevel(configuration.Logging.Level)

	var vault vault_helper.VaultHelper
	if configuration.Vault.Enabled {
		client, err := vault_helper.NewVaultHelper(log, vault_helper.Config{
			Address:         configuration.Vault.Address,
			Role:            configuration.Vault.Role,
			Auth:            configuration.Vault.Auth,
			Token:           configuration.Vault.Token,
			TokenFile:       configuration.Vault.TokenFile,
			Timeout:         30 * time.Second,
			IdleConnTimeout: 9 * time.Minute,
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

	pgOptions := &pg.Options{
		Addr: fmt.Sprintf("%s:%d",
			configuration.PostgreSQL.Address,
			configuration.PostgreSQL.Port,
		),
		User:            configuration.PostgreSQL.Username,
		Password:        configuration.PostgreSQL.Password,
		Database:        configuration.PostgreSQL.Database,
		ApplicationName: "monetr",
		MaxConnAge:      9 * time.Minute,
	}

	var tlsConfiguration *tls.Config

	if configuration.PostgreSQL.CACertificatePath != "" {
		pgOptions.MaxConnAge = 9 * time.Minute
		{
			caCert, err := ioutil.ReadFile(configuration.PostgreSQL.CACertificatePath)
			if err != nil {
				log.WithError(err).Errorf("failed to load ca certificate")
				return errors.Wrap(err, "failed to load ca certificate")
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfiguration = &tls.Config{
				Rand:               rand.Reader,
				InsecureSkipVerify: configuration.PostgreSQL.InsecureSkipVerify,
				RootCAs:            caCertPool,
				ServerName:         configuration.PostgreSQL.Address,
				Renegotiation:      tls.RenegotiateFreelyAsClient,
			}

			if configuration.PostgreSQL.KeyPath != "" {
				tlsCert, err := tls.LoadX509KeyPair(
					configuration.PostgreSQL.CertificatePath,
					configuration.PostgreSQL.KeyPath,
				)
				if err != nil {
					log.WithError(err).Errorf("failed to load client certificate")
					return errors.Wrap(err, "failed to load client certificate")
				}
				tlsConfiguration.Certificates = []tls.Certificate{
					tlsCert,
				}
			}
		}

		pgOptions.TLSConfig = tlsConfiguration
	}

	var db *pg.DB
	db = pg.Connect(pgOptions)
	db.AddQueryHook(logging.NewPostgresHooks(log, stats))
	pgOptions.OnConnect = func(ctx context.Context, cn *pg.Conn) error {
		log.Debugf("new connection with cert")

		return nil
	}

	if configuration.PostgreSQL.CACertificatePath != "" {
		paths := make([]string, 0, 1)
		for _, path := range []string{
			configuration.PostgreSQL.CACertificatePath,
			configuration.PostgreSQL.KeyPath,
			configuration.PostgreSQL.CertificatePath,
		} {
			directory := filepath.Dir(path)
			if !myownsanity.SliceContains(paths, directory) {
				paths = append(paths, directory)
			}
		}

		watchCertificate, err := certhelper.NewFileCertificateHelper(
			log,
			paths,
			func(path string) error {
				log.Info("reloading TLS certificates")

				tlsConfig := &tls.Config{
					Rand:               rand.Reader,
					InsecureSkipVerify: configuration.PostgreSQL.InsecureSkipVerify,
					RootCAs:            nil,
					ServerName:         configuration.PostgreSQL.Address,
					Renegotiation:      tls.RenegotiateFreelyAsClient,
				}

				{
					caCert, err := ioutil.ReadFile(configuration.PostgreSQL.CACertificatePath)
					if err != nil {
						log.WithError(err).Errorf("failed to load updated ca certificate")
						return errors.Wrap(err, "failed to load updated ca certificate")
					}

					caCertPool := x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(caCert)

					log.Debugf("new ca certificate loaded, swapping")

					tlsConfig.RootCAs = caCertPool
				}

				{
					if configuration.PostgreSQL.KeyPath != "" {
						tlsCert, err := tls.LoadX509KeyPair(
							configuration.PostgreSQL.CertificatePath,
							configuration.PostgreSQL.KeyPath,
						)
						if err != nil {
							log.WithError(err).Errorf("failed to load client certificate")
							return errors.Wrap(err, "failed to load client certificate")
						}

						tlsConfig.Certificates = []tls.Certificate{
							tlsCert,
						}
					}
				}

				db.Options().TLSConfig = tlsConfig

				log.Debugf("successfully swapped ca certificate")

				return nil
			},
		)
		if err != nil {
			log.WithError(err).Errorf("failed to setup certificate watcher")
			return errors.Wrap(err, "failed to setup certificate watcher")
		}
		watchCertificate.Start()

		defer watchCertificate.Stop()
	}

	if migrateDatabase {
		migrations.RunMigrations(log, db)
	} else {
		log.Info("automatic migrations are disabled")
	}

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
		log.Trace("stripe is enabled, creating client")
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
		plaidSecrets = secrets.NewVaultPlaidSecretsProvider(log, vault)
	} else {
		plaidSecrets = secrets.NewPostgresPlaidSecretsProvider(log, db)
	}

	plaidClient := platypus.NewPlaid(log, plaidSecrets, repository.NewPlaidRepository(db), configuration.Plaid)

	var smtpClient mail.Communication
	if configuration.Email.Enabled {
		smtpClient = mail.NewSMTPCommunication(log, configuration.Email.SMTP)
	}

	jobManager := jobs.NewJobManager(
		log,
		redisController.Pool(),
		db,
		plaidClient,
		stats,
		plaidSecrets,
	)
	defer jobManager.Close()

	app := application.NewApp(configuration, getControllers(
		log,
		configuration,
		db,
		jobManager,
		plaidClient,
		stats,
		stripe,
		redisController.Pool(),
		plaidSecrets,
		basicPaywall,
		smtpClient,
	)...)

	unixSocket := false

	idleConnsClosed := make(chan struct{})
	iris.RegisterOnInterrupt(func() {
		log.Info("shutting down")
		timeout := 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// close all hosts.
		_ = app.Shutdown(ctx)
		close(idleConnsClosed)
	})

	if unixSocket {
		workingDirectory, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		listener, err := net.ListenUnix("unix", &net.UnixAddr{
			Name: workingDirectory + "/api.sock",
			Net:  "unix",
		})
		if err != nil {
			panic(err)
		}

		return app.Run(iris.Listener(listener))
	} else {
		listenAddress := fmt.Sprintf(":%d", configuration.ListenPort)
		_ = app.Listen(listenAddress, iris.WithoutInterruptHandler, iris.WithoutServerError(iris.ErrServerClosed))
	}

	<-idleConnsClosed

	return nil
}
