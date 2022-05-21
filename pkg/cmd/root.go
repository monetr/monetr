package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCommand = &cobra.Command{
		Use:   "monetr",
		Short: "monetr's budgeting application",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	logLevelFlag string
)

func init() {
	rootCommand.PersistentFlags().StringVarP(&logLevelFlag, "log-level", "L", "info", "Specify the log level to use, allowed values: trace, debug, info, warn, error, fatal")
	newVersionCommand(rootCommand)
}
