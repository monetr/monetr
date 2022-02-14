package background

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	_ JobController = &SynchronousJobRunner{}
)

// SynchronousJobRunner is a harness around running jobs that allows jobs to be triggered normally outside this package
// but will run jobs synchronously. THIS IS MEANT FOR TESTING ONLY and should not be used in the actual monetr
// application.
type SynchronousJobRunner struct {
	t       *testing.T
	log     *logrus.Entry
	jobs    map[string]JobHandler
	marshal JobMarshaller
}

// NewSynchronousJobRunner will create a job runner for the current test. It does need to be provided the Platypus and
// PlaidSecretsProvider interfaces. But it will derive other requirements automatically, such as logs and the current
// database connection from the test context.
func NewSynchronousJobRunner(
	t *testing.T,
	plaidPlatypus platypus.Platypus,
	plaidSecrets secrets.PlaidSecretsProvider,
) *SynchronousJobRunner {
	if t == nil {
		panic("must be run within a test")
	}
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	runner := &SynchronousJobRunner{
		t:       t,
		log:     log,
		jobs:    map[string]JobHandler{},
		marshal: DefaultJobMarshaller,
	}

	publisher := pubsub.NewPostgresPubSub(log, db)

	jobs := []JobHandler{
		NewProcessFundingScheduleHandler(log, db),
		NewPullBalancesHandler(log, db, plaidSecrets, plaidPlatypus),
		NewPullTransactionsHandler(log, db, plaidSecrets, plaidPlatypus, publisher),
		NewRemoveLinkHandler(log, db, publisher),
		NewRemoveTransactionsHandler(log, db),
	}
	for i := range jobs {
		runner.jobs[jobs[i].QueueName()] = jobs[i]
	}

	return runner
}

func (s *SynchronousJobRunner) triggerJob(ctx context.Context, queue string, data interface{}) error {
	require.Contains(s.t, s.jobs, queue, "job must be registered in order to be triggered, might need to be updated?")
	jobHandler := s.jobs[queue]

	encodedArguments, err := s.marshal(data)
	require.NoError(s.t, err, "must be able to encode arguments for job")

	if err = jobHandler.HandleConsumeJob(ctx, encodedArguments); err != nil {
		s.log.WithContext(ctx).Error("synchronous job failure for test, this might not be desired behavior")
	}

	// A job failure would not return an error normally, it shouldn't return one here either.
	return nil
}
