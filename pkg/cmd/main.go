package main

import (
	"github.com/monetrapp/rest-api/pkg/build"
	"github.com/monetrapp/rest-api/pkg/internal/cmd"
	"log"
)

var (
	buildRevision = ""
	buildtime = ""
)

func main() {
	build.Revision = buildRevision
	build.BuildTime = buildtime
	// This is going to be the final actual program that is distributed.
	if err := cmd.RootCommand.Execute(); err != nil {
		log.Fatalf("failed: %+v", err)
	}
}
