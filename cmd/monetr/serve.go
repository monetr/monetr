package main

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/cache"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/controller"
	"github.com/monetrapp/rest-api/pkg/jobs"
	"github.com/monetrapp/rest-api/pkg/logging"
	"github.com/monetrapp/rest-api/pkg/metrics"
	"github.com/monetrapp/rest-api/pkg/migrations"
	"github.com/plaid/plaid-go/plaid"
	"github.com/spf13/cobra"
	"net/http"
)

func init() {
	rootCmd.AddCommand(serveCommand)
	serveCommand.LocalFlags().BoolVarP(&migrateDatabase, "migrate", "m", false, "Automatically run database migrations on startup. Defaults to: false")
	serveCommand.LocalFlags().StringVarP(&configFilePath, "config", "c", "", "Specify a config file to use, if omitted ./config.yaml or /etc/monetr/config.yaml will be used.")
}

var (
	configFilePath  = ""
	migrateDatabase = false

	serveCommand = &cobra.Command{
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

	log := logging.NewLogger()

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

	plaidClient, err := plaid.NewClient(plaid.ClientOptions{
		ClientID:    configuration.Plaid.ClientID,
		Secret:      configuration.Plaid.ClientSecret,
		Environment: configuration.Plaid.Environment,
		// TODO Don't use the default HTTP client for the Plaid client.
		HTTPClient: http.DefaultClient,
	})
	if err != nil {
		log.WithError(err).Fatalf("failed to create plaid client: %+v", err)
		return err
	}

	jobManager := jobs.NewJobManager(log, redisController.Pool(), db, plaidClient, stats)
	defer jobManager.Close()

	apiController := controller.NewController(configuration, db, jobManager, plaidClient, stats)

	app := application.NewApp(configuration, apiController)

	// TODO Allow listen port to be changed via config.
	return app.Listen(":4000")
}
