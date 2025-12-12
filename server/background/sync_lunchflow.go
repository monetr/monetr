package background

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/sirupsen/logrus"
)

const (
	SyncLunchFlow = "SyncLunchFlow"
)

var (
	_ ScheduledJobHandler = &SyncLunchFlowHandler{}
	_ JobImplementation   = &SyncLunchFlowJob{}
)

type (
	SyncLunchFlowHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		publisher    pubsub.Publisher
		enqueuer     JobEnqueuer
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	SyncLunchFlowArguments struct {
		AccountId     ID[Account]     `json:"accountId"`
		BankAccountId ID[BankAccount] `json:"bankAccountId"`
	}

	SyncLunchFlowJob struct {
		args      SyncLunchFlowArguments
		log       *logrus.Entry
		repo      repository.BaseRepository
		publisher pubsub.Publisher
		enqueuer  JobEnqueuer
		clock     clock.Clock
		timezone  *time.Location

		existingTransactions map[string]Transaction
	}
)

func NewSyncLunchFlowHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
) *SyncLunchFlowHandler {
	return &SyncLunchFlowHandler{
		log:          log,
		db:           db,
		publisher:    publisher,
		enqueuer:     enqueuer,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

// DefaultSchedule implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) DefaultSchedule() string {
	// Run every 12 hours at 30 minutes after the hour.
	return "0 30 */12 * * *"
}

// EnqueueTriggeredJob implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	panic("unimplemented")
}

// HandleConsumeJob implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) HandleConsumeJob(ctx context.Context, log *logrus.Entry, data []byte) error {
	panic("unimplemented")
}

// QueueName implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) QueueName() string {
	return SyncLunchFlow
}

// Run implements JobImplementation.
func (s *SyncLunchFlowJob) Run(ctx context.Context) error {
	panic("unimplemented")
}
