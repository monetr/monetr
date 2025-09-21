package main

import (
	"log"
	"runtime/debug"

	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/commands"
	"github.com/spf13/cobra"
)

var (
	buildRevision = ""
	buildTime     = ""
	buildHost     = ""
	buildType     = "binary"
	release       = ""
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
	commands.RegisterGlobalFlags(rootCommand)
	commands.AdminCommand(rootCommand)
	commands.DatabaseCommand(rootCommand)
	commands.DevelopmentCommand(rootCommand)
	commands.ServeCommand(rootCommand)
	commands.VersionCommand(rootCommand)
}

func main() {
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, item := range info.Settings {
			switch item.Key {
			case "vcs.revision":
				if item.Value != "" {
					buildRevision = item.Value
				}
			}
		}
	}
	build.Revision = buildRevision
	build.BuildTime = buildTime
	build.BuildHost = buildHost
	build.BuildType = buildType
	if release != "" {
		build.Release = release
	}
	// This is going to be the final actual program that is distributed.
	if err := rootCommand.Execute(); err != nil {
		log.Fatalf("failed: %+v", err)
	}
}
