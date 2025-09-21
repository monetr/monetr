package commands

import (
	"github.com/spf13/cobra"
)

func DatabaseCommand(parent *cobra.Command) {
	databaseCommand := &cobra.Command{
		Use:   "database",
		Short: "Manages the PostgreSQL database used by monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	databaseMigrate(databaseCommand)
	databaseVersion(databaseCommand)

	parent.AddCommand(databaseCommand)
}
