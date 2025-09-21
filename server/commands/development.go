//go:build development

package commands

import "github.com/spf13/cobra"

func DevelopmentCommand(parent *cobra.Command) {
	developmentCommand := &cobra.Command{
		Use:   "development",
		Short: "Development tools for working locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	developmentCacheFlush(developmentCommand)
	developmentCleanPlaid(developmentCommand)
	developmentCleanStripe(developmentCommand)

	parent.AddCommand(developmentCommand)
}
