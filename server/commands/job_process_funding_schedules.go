package commands

import (
	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/funding/funding_jobs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/queue"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func jobProcessFundingSchedules(parent *cobra.Command) {
	command := &cobra.Command{
		Use:   "process-funding-schedules",
		Short: "Trigger processing of all pending funding schedules.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clock := clock.New()
			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)

			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				return errors.Wrap(err, "failed to get database instance")
			}

			jobQueue := queue.NewPostgresQueue(
				cmd.Context(),
				clock,
				log,
				configuration,
				db,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)

			queueName := queue.QueueNameFromJobFunction[any](funding_jobs.ProcessFundingSchedulesCron)
			log.Info("enqueuing process funding schedules cron job", "queue", queueName)
			return jobQueue.EnqueueAt(cmd.Context(), queueName, clock.Now(), nil)
		},
	}

	parent.AddCommand(command)
}
