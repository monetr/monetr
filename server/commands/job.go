package commands

import "github.com/spf13/cobra"

func JobCommand(parent *cobra.Command) {
	jobCommand := &cobra.Command{
		Use:   "jobs [command] [flags]",
		Short: "Trigger jobs to be run by monetr instances or by this instance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	jobSyncPlaid(jobCommand)
	jobCleanupJobs(jobCommand)
	jobRemoveLink(jobCommand)

	parent.AddCommand(jobCommand)
}
