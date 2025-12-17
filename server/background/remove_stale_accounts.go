package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/sirupsen/logrus"
)

var (
	_ ScheduledJobHandler = &RemoveStaleAccountsHandler{}
	_ JobImplementation   = &RemoveStaleAccountsJob{}
)

const (
	RemoveStaleAccounts = "RemoveStaleAccounts"
)

type (
	RemoveStaleAccountsHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		configuration config.Configuration
		repo          repository.JobRepository
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	RemoveStaleAccountsArguments struct {
		AccountId ID[Account] `json:"accountId"`
	}

	RemoveStaleAccountsJob struct {
		args  RemoveStaleAccountsArguments
		log   *logrus.Entry
		repo  repository.BaseRepository
		clock clock.Clock
	}
)

func NewRemoveStaleAccountsHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	configuration config.Configuration,
) *RemoveStaleAccountsHandler {
	return &RemoveStaleAccountsHandler{
		log:           log,
		db:            db,
		configuration: configuration,
		repo:          repository.NewJobRepository(db, clock),
		unmarshaller:  DefaultJobUnmarshaller,
		clock:         clock,
	}
}

// DefaultSchedule implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) DefaultSchedule() string {
	panic("unimplemented")
}

// EnqueueTriggeredJob implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	panic("unimplemented")
}

// HandleConsumeJob implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) HandleConsumeJob(ctx context.Context, log *logrus.Entry, data []byte) error {
	panic("unimplemented")
}

// QueueName implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) QueueName() string {
	panic("unimplemented")
}

// Run implements JobImplementation.
func (r *RemoveStaleAccountsJob) Run(ctx context.Context) error {
	panic("unimplemented")
}
