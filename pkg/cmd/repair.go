package main

import "github.com/spf13/cobra"

func init() {
	rootCommand.AddCommand(RepairCommand)
	RepairCommand.AddCommand(RepairPlaidCommand)

	RepairPlaidCommand.AddCommand(RepairPlaidWebhooksCommand)
}

var (
	RepairCommand = &cobra.Command{
		Use:   "repair",
		Short: "Can repair or update some inconsistencies in the database when major config changes have occurred.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	RepairPlaidCommand = &cobra.Command{
		Use:   "plaid",
		Short: "Plaid repair tasks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	RepairPlaidWebhooksCommand = &cobra.Command{
		Use:   "webhooks",
		Short: "Will update the webhooks URL for all Plaid links in the database that do not match the currently configured webhook URL.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)
