package background

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewBackgroundJobs(t *testing.T) {
	t.Run("invalid background engine", func(t *testing.T) {
		clock := clock.NewMock()
		ctx := context.Background()
		log := testutils.GetLog(t)
		configuration := config.Configuration{
			BackgroundJobs: config.BackgroundJobs{
				Engine:      "ironmq", // Not a valid engine.
				Scheduler:   config.BackgroundJobSchedulerExternal,
				JobSchedule: nil,
			},
		}

		jobs, err := NewBackgroundJobs(ctx, log, clock, configuration, nil, nil, nil, nil, nil, nil)
		assert.Nil(t, jobs, "object returned should be nil")
		assert.EqualError(t, err, "invalid background job engine specified")
	})
}
