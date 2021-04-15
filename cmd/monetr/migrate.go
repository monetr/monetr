package main

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(databaseCommand)
	databaseCommand.AddCommand(migrateCommand)
	databaseCommand.AddCommand(databaseVersionCommand)

	databaseCommand.PersistentFlags().StringVarP(&postgresAddress, "host", "H", "", "PostgreSQL host address.")
	databaseCommand.PersistentFlags().IntVarP(&postgresPort, "port", "P", 5432, "PostgreSQL port.")
}

var (
	postgresAddress  = ""
	postgresPort     = 5432
	postgresUsername = "postgres"
	postgresPassword = ""
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
