package commands

import "github.com/spf13/cobra"

func AdminCommand(parent *cobra.Command) {
	adminCommand := &cobra.Command{
		Use:   "admin",
		Short: "General administrative tasks for hosting/maintaining monetr",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	adminKMSCheck(adminCommand)
	adminKMSMigrate(adminCommand)
	adminPlaidRefresh(adminCommand)
	adminRegisterCode(adminCommand)
	adminSecretView(adminCommand)

	parent.AddCommand(adminCommand)
}
