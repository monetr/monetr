package main

import (
	"fmt"

	"github.com/monetr/monetr/pkg/build"
	"github.com/spf13/cobra"
)

func newNoticesCommand(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "notices",
		Short: "Prints the third party notices to STDOUT.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(build.GetNotice())
			return nil
		},
	}

	parent.AddCommand(command)
}
