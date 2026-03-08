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
				log.Info("config file loaded", "config", configFileName)
			}

			redisController, err := cache.NewRedisCache(log, configuration.Redis)
			if err != nil {
				log.Error("failed to create redis cache", "err", err)
				return err
			}
			defer redisController.Close()

			conn, err := redisController.Pool().Dial()
			if err != nil {
				log.Error("failed to retrieve connection from redis pool", "err", err)
				return err
			}

			if err := conn.Send("FLUSHALL"); err != nil {
				log.Error("failed to flush redis cache", "err", err)
				return err
			}

			log.Info("done!")
			return nil
		},
	}

	parent.AddCommand(command)
}
