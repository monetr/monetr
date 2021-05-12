package cmd

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/build"
	"github.com/monetrapp/rest-api/pkg/cache"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/internal/migrations"
	"github.com/monetrapp/rest-api/pkg/internal/plaid_helper"
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/logging"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v72"
	stripe_client "github.com/stripe/stripe-go/v72/client"
	"net"
	"net/http"
	"os"
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
		// TODO Add support for TLS with PostgreSQL.
	}

	db := pg.Connect(pgOptions)
	db.AddQueryHook(logging.NewPostgresHooks(log, stats))

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
