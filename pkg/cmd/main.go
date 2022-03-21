package main

import (
	"log"

	"github.com/monetr/monetr/pkg/build"
)

var (
	buildRevision = ""
	buildTime     = ""
	buildHost     = ""
	release       = ""
)

func main() {
	build.Revision = buildRevision
	build.BuildTime = buildTime
	build.BuildHost = buildHost
	if release == "" {
		build.Release = buildRevision
	} else {
		build.Release = release
	}
	// This is going to be the final actual program that is distributed.
	if err := rootCommand.Execute(); err != nil {
		log.Fatalf("failed: %+v", err)
	}
}
