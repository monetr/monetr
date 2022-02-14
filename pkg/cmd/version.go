package main

import (
	"fmt"
	"runtime"

	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/ui"
	"github.com/spf13/cobra"
)

type versionCommand struct {
	detailed bool
}

func newVersionCommand(parent *cobra.Command) {
	var arguments versionCommand
	command := &cobra.Command{
		Use:   "version",
		Short: "Print the version of monetr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !arguments.detailed {
				fmt.Println(build.Release)
				return nil
			}

			detailedString := "" +
				"Version:      %s\n" +
				"Revision:     %s\n" +
				"Build time:   %s\n" +
				"Embedded UI:  %t\n" +
				"Architecture: %s\n" +
				"OS:           %s\n" +
				"Compiler:     %s\n"

			fmt.Printf(
				detailedString,
				build.Release,
				build.Revision,
				build.BuildTime,
				ui.EmbeddedUI,
				runtime.GOARCH,
				runtime.GOOS,
				runtime.Compiler,
			)

			return nil
		},
	}
	command.PersistentFlags().BoolVarP(&arguments.detailed, "detailed", "d", false, "Print detailed version information.")

	parent.AddCommand(command)
}
