package main

import (
	"github.com/monetr/monetr/pkg/config"
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
)

func init() {
	rootCommand.PersistentFlags().StringVarP(&config.LogLevel, "log-level", "L", "info", "Specify the log level to use, allowed values: trace, debug, info, warn, error, fatal")
	rootCommand.PersistentFlags().StringVarP(&config.FilePath, "config", "c", "", "Specify the config file to use.")
	newVersionCommand(rootCommand)
}
