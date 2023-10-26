package main

import (
	"log"
	"runtime/debug"

	"github.com/monetr/monetr/server/build"
)

var (
	buildRevision = ""
	buildTime     = ""
	buildHost     = ""
	buildType     = "binary"
	release       = ""
)

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
