package jobs

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	CleanupJobsTable = "CleanupJobsTable"
)

func (j *jobManagerBase) cleanupJobsTable(job *work.Job) (err error) {
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Pull Account Balances"))
	defer span.Finish()

	defer j.recover(span.Context())

	defer func() {
		if err := recover(); err != nil {
			sentry.CaptureException(errors.Errorf("pull account balances failure: %+v", err))
		}
	}()

	log := j.getLogForJob(job)
	log.Infof("getting ready to clean up the jobs table")

	result, err := j.db.ModelContext(span.Context(), &models.Job{}).
		Where(`"job"."enqueued_at" < ?`, time.Now().Add(-15*24*time.Hour)).
		Delete()
	if err != nil {
		log.WithError(err).Errorf("failed to cleanup old jobs from the jobs table")
		return errors.Wrap(err, "failed to cleanup old jobs from the jobs table")
	}

	if affected := result.RowsAffected(); affected > 0 {
		log.WithFields(logrus.Fields{
			"deleted": affected,
		}).Infof("deleted old jobs from the jobs table")
	} else {
		log.Infof("no jobs were cleaned up from the jobs table")
	}

	return nil
}
