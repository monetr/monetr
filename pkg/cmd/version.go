package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/icons"
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
				"Version:         %s\n" +
				"Revision:        %s\n" +
				"Build time:      %s\n" +
				"Build host:      %s\n" +
				"Embedded UI:     %t\n" +
				"Embedded Icons:  %t\n" +
				"  Icon Packs:    %s\n" +
				"Architecture:    %s\n" +
				"OS:              %s\n" +
				"Compiler:        %s\n" +
				"Go Version:      %s\n"

			iconsEnabled := icons.GetIconsEnabled()
			iconPacks := "<not enabled>"
			if iconsEnabled {
				indexes := icons.GetIconIndexes()
				if len(indexes) == 0 {
					iconPacks = "<none enabled>"
				} else {
					iconPacks = strings.Join(indexes, ", ")
				}
			}

			fmt.Printf(
				detailedString,
				build.Release,
				build.Revision,
				build.BuildTime,
				build.BuildHost,
				ui.EmbeddedUI,
				iconsEnabled,
				iconPacks,
				runtime.GOARCH,
				runtime.GOOS,
				runtime.Compiler,
				runtime.Version(),
			)

			return nil
		},
	}
	command.PersistentFlags().BoolVarP(&arguments.detailed, "detailed", "d", false, "Print detailed version information.")

	parent.AddCommand(command)
}
