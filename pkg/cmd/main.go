package main

import (
	"log"

	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/internal/cmd"
)

var (
	buildRevision = ""
	buildtime     = ""
	release       = ""
)

func main() {
	build.Revision = buildRevision
	build.BuildTime = buildtime
	if release == "" {
		build.Release = buildRevision
	} else {
		build.Release = release
	}
	// This is going to be the final actual program that is distributed.
	if err := cmd.RootCommand.Execute(); err != nil {
		log.Fatalf("failed: %+v", err)
	}
}
