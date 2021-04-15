package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(databaseCommand)
	databaseCommand.AddCommand(migrateCommand)
	databaseCommand.AddCommand(databaseVersionCommand)

	databaseCommand.PersistentFlags().StringVarP(&postgresAddress, "host", "H", "", "PostgreSQL host address.")
	databaseCommand.PersistentFlags().IntVarP(&postgresPort, "port", "P", 0, "PostgreSQL port.")
	databaseCommand.PersistentFlags().StringVarP(&postgresUsername, "username", "U", "", "PostgreSQL user.")
	databaseCommand.PersistentFlags().StringVarP(&postgresPassword, "password", "W", "", "PostgreSQL password.")
	databaseCommand.PersistentFlags().StringVarP(&postgresDatabase, "database", "d", "", "PostgreSQL database.")
}

var (
	postgresAddress  = ""
	postgresPort     = 0
	postgresUsername = ""
	postgresPassword = ""
	postgresDatabase = ""
)

var (
	migrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations against your PostgreSQL.",
		Long:  "Updates your PostgreSQL database to the latest schema version for monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunMigration()
		},
	}

	databaseVersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Prints version information about your database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunMigration()
		},
	}

	databaseCommand = &cobra.Command{
		Use:   "database",
		Short: "Manages the PostgreSQL database used by monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunMigration()
		},
	}
)

func RunMigration() error {
	return nil
}
