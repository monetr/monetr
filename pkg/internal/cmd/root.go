package cmd

import "github.com/spf13/cobra"

var (
	RootCommand = &cobra.Command{
		Use:   "monetr",
		Short: "REST-API for monetr's budgeting application",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)
