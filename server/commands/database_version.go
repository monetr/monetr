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
				log.Error("failed to setup database", "err", err)
				return err
			}

			migrator, err := migrations.NewMigrationsManager(cmd.Context(), log, db)
			if err != nil {
				log.Error("failed to create migration manager", "err", err)
				return err
			}

			fmt.Println("Latest:", migrator.LatestVersion())

			version, err := migrator.CurrentVersion(cmd.Context())
			if err != nil {
				log.Error("failed to determine current database version", "err", err)
				return err
			}

			// No logging frills, just print the version to STDOUT
			fmt.Println("Current:", version)

			return nil
		},
	}

	parent.AddCommand(command)
}
