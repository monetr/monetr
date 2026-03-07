package commands

import (
	"fmt"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/migrations"
	"github.com/spf13/cobra"
)

func databaseMigrate(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations against your PostgreSQL.",
		Long:  "Updates your PostgreSQL database to the latest schema version for monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			// Overwrite this value since we are managing the migration ourselves.
			configuration.PostgreSQL.Migrate = false
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.Info("config file loaded", "config", configFileName)
			}
			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.Error("failed to establish database connection", "err", err)
				return err
			}
			defer db.Close()

			migrator, err := migrations.NewMigrationsManager(log, db)
			if err != nil {
				log.Error("failed to create migration manager", "err", err)
				return err
			}

			oldVersion, newVersion, err := migrator.Up()
			if err != nil {
				log.Error("failed to run schema migrations", "err", err)
				return err
			}

			if oldVersion != newVersion {
				log.Info(fmt.Sprintf("successfully upgraded database from %d to %d", oldVersion, newVersion))
			} else {
				log.Info("database is up to date, no migrations were run")
			}

			return nil
		},
	}

	parent.AddCommand(command)
}
