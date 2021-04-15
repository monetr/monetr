package main

import (
	"github.com/spf13/cobra"
	"log"
)

func main() {
	// This is going to be the final actual program that is distributed.
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed: %+v", err)
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "monetr",
		Short: "REST-API for monetr's budgeting application",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)
