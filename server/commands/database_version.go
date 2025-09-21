package commands

import (
	"fmt"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/migrations"
	"github.com/spf13/cobra"
)

func databaseVersion(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "version",
		Short: "Prints version information about your database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			configuration.PostgreSQL.Migrate = false
			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatal("failed to setup database")
			}

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

	parent.AddCommand(command)
}
