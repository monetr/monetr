package background

import (
	"context"
	"fmt"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ JobHandler        = &RemoveAccountHandler{}
	_ JobImplementation = &RemoveAccountJob{}
)

const (
	RemoveAccount = "RemoveAccount"
)

func TriggerRemoveAccount(
	ctx context.Context,
	backgroundJobs JobController,
	arguments SyncPlaidArguments,
) error {
	if arguments.Trigger == "" {
		arguments.Trigger = "manual"
	}
	return backgroundJobs.EnqueueJob(ctx, SyncPlaid, arguments)
}

type (
	RemoveAccountHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		configuration config.Configuration
		kms           secrets.KeyManagement
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	RemoveAccountArguments struct {
		AccountId ID[Account] `json:"accountId"`
	}

	RemoveAccountJob struct {
		args    RemoveAccountArguments
		log     *logrus.Entry
		repo    repository.BaseRepository
		secrets repository.SecretsRepository
		clock   clock.Clock
	}
)

func (r *RemoveAccountHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args RemoveAccountArguments
	if err := errors.Wrap(r.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Remove Account job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return r.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context()).WithFields(logrus.Fields{
			"accountId": args.AccountId,
		})
		log.Trace("placeholder")

		return nil
	})
}

func (r *RemoveAccountHandler) QueueName() string {
	return RemoveAccount
}

// Run implements JobImplementation.
func (r *RemoveAccountJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := r.log.WithContext(span.Context())

	log.Trace("checking for still active Plaid links before account can be removed")
	allLinks, err := r.repo.GetLinks(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to read links to check for active plaid links")
	}
	fmt.Sprint(allLinks)

	// { // Check for active plaid links
	// 	for _, link := range allLinks {
	// 		if
	// 	}
	// }

	return nil
}
