package commands

import "github.com/spf13/cobra"

func AdminCommand(parent *cobra.Command) {
	adminCommand := &cobra.Command{
		Use:   "admin",
		Short: "General administrative tasks for hosting/maintaining monetr",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	adminKMSCheck(adminCommand)
	adminKMSMigrate(adminCommand)
	adminLoginResetPassword(adminCommand)
	adminPlaidAccounts(adminCommand)
	adminPlaidRefresh(adminCommand)
	adminPlaidUpdateWebhook(adminCommand)
	adminRegisterCode(adminCommand)

	parent.AddCommand(adminCommand)
}
