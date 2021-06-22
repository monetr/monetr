package cmd

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/build"
	"github.com/monetrapp/rest-api/pkg/cache"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/internal/certhelper"
	"github.com/monetrapp/rest-api/pkg/internal/migrations"
	"github.com/monetrapp/rest-api/pkg/internal/myownsanity"
	"github.com/monetrapp/rest-api/pkg/internal/plaid_helper"
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/logging"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v72"
	stripe_client "github.com/stripe/stripe-go/v72/client"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
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
						delete(event.Request.Headers, "M-Token")
						delete(event.Request.Headers, "Cookie")
						delete(event.Request.Headers, "Cookies")
					}
				}

				return event
			},
		})
		if err != nil {
			log.WithError(err).Error("failed to init sentry")
		}
		defer sentry.Flush(10 * time.Second)
	}

	var stripeClient *stripe_client.API
	if configuration.Stripe.Enabled {
		log.Trace("stripe is enabled, creating client")
		stripeClient = stripe_client.New(configuration.Stripe.APIKey, stripe.NewBackends(http.DefaultClient))
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
				InsecureSkipVerify: false,
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
					InsecureSkipVerify: false,
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

				//newPointer := unsafe.Pointer(tlsConfig)
				//existingPointer := (*unsafe.Pointer)(unsafe.Pointer(tlsConfiguration))
				//atomic.SwapPointer(existingPointer, newPointer)

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

	plaidHelper := plaid_helper.NewPlaidClient(log, plaid.ClientOptions{
		ClientID:    configuration.Plaid.ClientID,
		Secret:      configuration.Plaid.ClientSecret,
		Environment: configuration.Plaid.Environment,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})

	if configuration.Plaid.WebhooksEnabled {
		log.Debugf("plaid webhooks are enabled and will be sent to: %s", configuration.Plaid.WebhooksDomain)
	}

	if configuration.Stripe.Enabled && configuration.Stripe.WebhooksEnabled {
		log.Debugf("stripe webhooks are enabled and will be sent to: %s", configuration.Stripe.WebhooksDomain)
	}

	jobManager := jobs.NewJobManager(log, redisController.Pool(), db, plaidHelper, stats)
	defer jobManager.Close()

	app := application.NewApp(configuration, getControllers(
		log,
		configuration,
		db,
		jobManager,
		plaidHelper,
		stats,
		stripeClient,
		redisController.Pool(),
	)...)

	unixSocket := false

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
		// TODO Allow listen port to be changed via config.
		return app.Listen(":4000")
	}
}
