package main

import (
	"github.com/monetrapp/rest-api/pkg/internal/cmd"
	"log"
)

func main() {
	// This is going to be the final actual program that is distributed.
	if err := cmd.RootCommand.Execute(); err != nil {
		log.Fatalf("failed: %+v", err)
	}
}
