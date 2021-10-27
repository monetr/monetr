package cmd

import (
	"fmt"

	"github.com/monetr/monetr/pkg/build"
	"github.com/spf13/cobra"
)

var (
	RootCommand = &cobra.Command{
		Use:   "monetr",
		Short: "monetr's budgeting application",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	VersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Print the version of monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(build.Release)
			return nil
		},
	}
)

func init() {
	RootCommand.AddCommand(VersionCommand)
}
