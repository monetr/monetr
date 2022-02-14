package main

import (
	"github.com/monetr/monetr/pkg/background"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/logging"
	"github.com/spf13/cobra"
)

func newCleanupJobsCommand(parent *cobra.Command) {
	var dryRun bool
	var local bool

	command := &cobra.Command{
		Use:   "cleanup-jobs",
		Short: "Cleanup old job records in the job table.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var configPath *string
			if len(configFilePath) > 0 {
				configPath = &configFilePath
			}

			configuration := config.LoadConfiguration(configPath)
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := getDatabase(log, configuration, nil)

			if local || dryRun {
				log.Info("running locally")

				txn, err := db.BeginContext(cmd.Context())
				if err != nil {
					log.WithError(err).Fatalf("failed to begin transaction to cleanup jobs")
					return err
				}

				job := background.NewCleanupJobsJob(log, txn)
				if err := job.Run(cmd.Context()); err != nil {
					log.WithError(err).Fatalf("failed to run cleanup jobs")
					_ = txn.RollbackContext(cmd.Context())
					return err
				}

				if dryRun {
					log.Info("dry run... rolling changes back")
					return txn.RollbackContext(cmd.Context())
				} else {
					return txn.CommitContext(cmd.Context())
				}
			}

			redisController, err := cache.NewRedisCache(log, configuration.Redis)
			if err != nil {
				log.WithError(err).Fatalf("failed to create redis cache: %+v", err)
				return err
			}
			defer redisController.Close()

			backgroundJobs, err := background.NewBackgroundJobs(
				cmd.Context(),
				log,
				configuration,
				db,
				redisController.Pool(),
				nil,
				nil,
				nil,
			)
			if err != nil {
				return err
			}

			return background.TriggerCleanupJobs(cmd.Context(), backgroundJobs)
		},
	}

	command.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run the job cleanup, this will log some basic information about what records would be removed but will not persist any changes. [local]")
	command.PersistentFlags().BoolVar(&local, "local", false, "Run the job locally, this means the job is not dispatched to the external scheduler like RabbitMQ or Redis. This defaults to true when dry running or when the job engine is in-memory.")
	parent.AddCommand(command)
}
