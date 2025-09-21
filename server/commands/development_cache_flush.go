//go:build development

package commands

import (
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/logging"
	"github.com/spf13/cobra"
)

func developmentCacheFlush(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "cache:flush",
		Short: "Flush all data from the Redis cache server.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configuration := config.LoadConfiguration()

			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.WithField("config", configFileName).Info("config file loaded")
			}

			redisController, err := cache.NewRedisCache(log, configuration.Redis)
			if err != nil {
				log.WithError(err).Fatalf("failed to create redis cache: %+v", err)
				return err
			}
			defer redisController.Close()

			conn, err := redisController.Pool().Dial()
			if err != nil {
				log.WithError(err).Fatalf("failed to retrieve connection from redis pool: %+v", err)
				return err
			}

			if err := conn.Send("FLUSHALL"); err != nil {
				log.WithError(err).Fatalf("failed to flush redis cache: %+v", err)
				return err
			}

			log.Info("done!")
			return nil
		},
	}

	parent.AddCommand(command)
}
