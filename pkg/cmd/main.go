package main

import (
	"log"

	"github.com/monetr/rest-api/pkg/build"
	"github.com/monetr/rest-api/pkg/internal/cmd"
)

var (
	buildRevision = ""
	buildtime     = ""
	release       = ""
)

func main() {
	build.Revision = buildRevision
	build.BuildTime = buildtime
	build.Release = release
	// This is going to be the final actual program that is distributed.
	if err := cmd.RootCommand.Execute(); err != nil {
		log.Fatalf("failed: %+v", err)
	}
}
