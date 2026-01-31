package main

import (
	"context"
	"time"

	"github.com/monetr/monetr/server/backup"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/spf13/cobra"
)

func init() {
	newBackupCommand(rootCommand)
}

func newBackupCommand(parent *cobra.Command) {
	var path string = "/Users/elliotcourant/.monetr"
	var kind string = "filesystem"
	var chunkSize int = 1024 * 1024 * 1 // 1MB
	command := &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of your monetr data",
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}
			configuration.Backup.Prefix = path
			configuration.Backup.Kind = kind
			configuration.Backup.ChunkSize = chunkSize

			db, err := getDatabase(log, configuration, nil)
			if err != nil {
				log.WithError(err).Fatal("failed to setup database")
				return err
			}

			kms, err := getKMS(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("failed to initialize KMS")
				return err
			}

			fileStorage, err := setupStorage(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("could not setup file storage")
				return err
			}

			destination, err := backup.NewBackupDestination(log, configuration)
			if err != nil {
				log.WithError(err).Fatal("could not setup backup destination")
				return err
			}

			manager := backup.NewBackupManager(
				log,
				configuration,
				db,
				kms,
				fileStorage,
				destination,
			)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			err = manager.Backup(ctx)
			return err
		},
	}

	parent.AddCommand(command)
}
