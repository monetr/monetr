package background

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewBackgroundJobs(t *testing.T) {
	t.Run("panics for rabbitmq", func(t *testing.T) {
		ctx := context.Background()
		log := testutils.GetLog(t)
		configuration := config.Configuration{
			BackgroundJobs: config.BackgroundJobs{
				Engine:      config.BackgroundJobEngineRabbitMQ,
				Scheduler:   config.BackgroundJobSchedulerExternal,
				JobSchedule: nil,
			},
		}

		var jobs *BackgroundJobs
		var err error
		assert.Panics(t, func() {
			jobs, err = NewBackgroundJobs(ctx, log, configuration, nil, nil, nil, nil, nil)
		}, "must panic if rabbitmq is specified")

		assert.Nil(t, jobs, "object returned should be nil")
		assert.NoError(t, err, "must not return an error")
	})

	t.Run("invalid background engine", func(t *testing.T) {
		ctx := context.Background()
		log := testutils.GetLog(t)
		configuration := config.Configuration{
			BackgroundJobs: config.BackgroundJobs{
				Engine:      "ironmq", // Not a valid engine.
				Scheduler:   config.BackgroundJobSchedulerExternal,
				JobSchedule: nil,
			},
		}

		jobs, err := NewBackgroundJobs(ctx, log, configuration, nil, nil, nil, nil, nil)
		assert.Nil(t, jobs, "object returned should be nil")
		assert.EqualError(t, err, "invalid background job engine specified")
	})
}
