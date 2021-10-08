package jobs

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRemoveLinkJob_Run(t *testing.T) {
	t.Run("does nothing", func(t *testing.T) {
		job := &RemoveLinkJob{
			db:  testutils.GetPgDatabase(t),
			log: testutils.GetLog(t),
		}

		err := job.Run(context.Background())
		assert.NoError(t, err, "this job has no link Id and should do nothing")
	})
}
