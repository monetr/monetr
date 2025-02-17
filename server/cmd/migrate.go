package main

import (
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/migrations"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCommand.AddCommand(DatabaseCommand)
	DatabaseCommand.AddCommand(MigrateCommand)
	DatabaseCommand.AddCommand(DatabaseVersionCommand)

	DatabaseCommand.PersistentFlags().StringVarP(&postgresAddress, "host", "H", "", "PostgreSQL host address.")
	DatabaseCommand.PersistentFlags().IntVarP(&postgresPort, "port", "P", 0, "PostgreSQL port.")
	DatabaseCommand.PersistentFlags().StringVarP(&postgresUsername, "username", "U", "", "PostgreSQL user.")
	DatabaseCommand.PersistentFlags().StringVarP(&postgresPassword, "password", "W", "", "PostgreSQL password.")
	DatabaseCommand.PersistentFlags().StringVarP(&postgresDatabase, "database", "d", "", "PostgreSQL database.")

	// TODO This doesn't account for TLS properties that would need to be set.
	viper.BindPFlag("PostgreSQL.Address", DataCommand.PersistentFlags().Lookup("host"))
	viper.BindPFlag("PostgreSQL.Port", DataCommand.PersistentFlags().Lookup("port"))
	viper.BindPFlag("PostgreSQL.Username", DataCommand.PersistentFlags().Lookup("username"))
	viper.BindPFlag("PostgreSQL.Password", DataCommand.PersistentFlags().Lookup("password"))
	viper.BindPFlag("PostgreSQL.Database", DataCommand.PersistentFlags().Lookup("database"))
}

var (
	postgresAddress  = ""
	postgresPort     = 0
	postgresUsername = ""
	postgresPassword = ""
	postgresDatabase = ""
)

var (
	MigrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations against your PostgreSQL.",
		Long:  "Updates your PostgreSQL database to the latest schema version for monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			// Overwrite this value since we are managing the migration ourselves.
			configuration.PostgreSQL.Migrate = false
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}
			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatalf("failed to establish database connection")
				return err
			}
			defer db.Close()

			migrator, err := migrations.NewMigrationsManager(log, db)
			if err != nil {
				log.WithError(err).Fatalf("failed to create migration manager")
				return err
			}

			oldVersion, newVersion, err := migrator.Up()
			if err != nil {
				log.WithError(err).Fatalf("failed to run schema migrations")
				return err
			}

			if oldVersion != newVersion {
				log.Infof("successfully upgraded database from %d to %d", oldVersion, newVersion)
			} else {
				log.Info("database is up to date, no migrations were run")
			}

			return nil
		},
	}

	DatabaseVersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Prints version information about your database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logging.NewLoggerWithLevel(config.LogLevel)

			options := getDatabaseCommandConfiguration()

			db := pg.Connect(options)

			migrator, err := migrations.NewMigrationsManager(log, db)
			if err != nil {
				log.WithError(err).Fatalf("failed to create migration manager")
				return err
			}

			latestVersion, err := migrator.LatestVersion()
			if err != nil {
				log.WithError(err).Fatalf("failed to determine latest database version")
				return err
			}

			fmt.Println("Latest:", latestVersion)

			version, err := migrator.CurrentVersion()
			if err != nil {
				log.WithError(err).Fatalf("failed to determine current database version")
				return err
			}

			// No logging frills, just print the version to STDOUT
			fmt.Println("Current:", version)

			return nil
		},
	}

	DatabaseCommand = &cobra.Command{
		Use:   "database",
		Short: "Manages the PostgreSQL database used by monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func getDatabaseCommandConfiguration() *pg.Options {
	configuration := config.LoadConfiguration()

	address := myownsanity.CoalesceStrings(postgresAddress, configuration.PostgreSQL.Address, "localhost")
	port := myownsanity.CoalesceInts(postgresPort, configuration.PostgreSQL.Port, 5432)
	username := myownsanity.CoalesceStrings(postgresUsername, configuration.PostgreSQL.Username, "postgres")
	password := myownsanity.CoalesceStrings(postgresPassword, configuration.PostgreSQL.Password)
	database := myownsanity.CoalesceStrings(postgresDatabase, configuration.PostgreSQL.Database, "postgres")

	options := &pg.Options{
		Addr:            fmt.Sprintf("%s:%d", address, port),
		User:            username,
		Password:        password,
		Database:        database,
		ApplicationName: "monetr",
	}

	return options
}
