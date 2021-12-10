package jobs

import (
	"context"
	"math"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/internal/platypus"
	"github.com/uptrace/bun"

	"github.com/go-pg/pg/v10"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/pkg/metrics"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type JobManager interface {
	TriggerPullHistoricalTransactions(accountId, linkId uint64) (jobId string, err error)
	TriggerPullInitialTransactions(accountId, userId, linkId uint64) (jobId string, err error)
	TriggerPullLatestTransactions(accountId, linkId uint64, numberOfTransactions int64) (jobId string, err error)
	TriggerRemoveTransactions(accountId, linkId uint64, removedTransactions []string) (jobId string, err error)
	TriggerRemoveLink(accountId, userId, linkId uint64) (jobId string, err error)
	Close() error
}

var (
	_ JobManager = &jobManagerBase{}
	_ JobManager = &nonDistributedJobManager{}
)

type jobManagerBase struct {
	log          *logrus.Entry
	work         *work.WorkerPool
	queue        *work.Enqueuer
	db           *bun.DB
	plaidClient  platypus.Platypus
	plaidSecrets secrets.PlaidSecretsProvider
	stats        *metrics.Stats
	ps           pubsub.PublishSubscribe
}

func NewNonDistributedJobManager(
	log *logrus.Entry,
	pool *redis.Pool,
	db *bun.DB,
	plaidClient platypus.Platypus,
	stats *metrics.Stats,
	plaidSecrets secrets.PlaidSecretsProvider,
) JobManager {
	manager := &nonDistributedJobManager{
		log: log,
		// TODO (elliotcourant) Use namespace from config.
		db:           db,
		plaidClient:  plaidClient,
		plaidSecrets: plaidSecrets,
		stats:        stats,
		ps:           pubsub.NewPostgresPubSub(log, db),
	}

	return manager
}

func NewJobManager(
	log *logrus.Entry,
	pool *redis.Pool,
	db *pg.DB,
	plaidClient platypus.Platypus,
	stats *metrics.Stats,
	plaidSecrets secrets.PlaidSecretsProvider,
) JobManager {
	manager := &jobManagerBase{
		log: log,
		// TODO (elliotcourant) Use namespace from config.
		work:         work.NewWorkerPool(struct{}{}, 4, "harder", pool),
		queue:        work.NewEnqueuer("harder", pool),
		db:           db,
		plaidClient:  plaidClient,
		plaidSecrets: plaidSecrets,
		stats:        stats,
		ps:           pubsub.NewPostgresPubSub(log, db),
	}

	manager.work.Middleware(manager.middleware)

	manager.work.Job(EnqueueProcessFundingSchedules, manager.enqueueProcessFundingSchedules)
	manager.work.Job(EnqueuePullAccountBalances, manager.enqueuePullAccountBalances)
	manager.work.Job(EnqueuePullLatestTransactions, manager.enqueuePullLatestTransactions)

	manager.work.Job(CleanupJobsTable, manager.cleanupJobsTable)
	manager.work.Job(ProcessFundingSchedules, manager.processFundingSchedules)
	manager.work.Job(PullAccountBalances, manager.pullAccountBalances)
	manager.work.Job(PullHistoricalTransactions, manager.pullHistoricalTransactions)
	manager.work.Job(PullInitialTransactions, manager.pullInitialTransactions)
	manager.work.Job(PullLatestTransactions, manager.pullLatestTransactions)
	manager.work.Job(RemoveLink, manager.removeLink)
	manager.work.Job(RemoveTransactions, manager.removeTransactions)

	// Every 30 minutes. 0 */30 * * * *

	// Every hour.
	manager.work.PeriodicallyEnqueue("0 0 * * * *", EnqueueProcessFundingSchedules)

	// Once a day. But also can be triggered by a webhook.
	// Once A day. 0 0 0 * * *
	manager.work.PeriodicallyEnqueue("0 0 0 * * *", EnqueuePullAccountBalances)
	manager.work.PeriodicallyEnqueue("0 0 0 * * *", EnqueuePullLatestTransactions)

	// Once a day at 8AM
	manager.work.PeriodicallyEnqueue("0 0 8 * * *", CleanupJobsTable)

	manager.work.Start()
	log.Debug("job manager started")

	return manager
}

func (j *jobManagerBase) enqueueUniqueJob(name string, arguments map[string]interface{}) (*work.Job, error) {
	if j.stats != nil {
		j.stats.JobEnqueued(name)
	}

	return j.queue.EnqueueUnique(name, arguments)
}

func (j *jobManagerBase) TriggerPullInitialTransactions(accountId, userId, linkId uint64) (jobId string, err error) {
	job, err := j.queue.EnqueueUnique(PullInitialTransactions, map[string]interface{}{
		"accountId": accountId,
		"userId":    userId,
		"linkId":    linkId,
	})
	if err != nil {
		return "", err
	}

	return job.ID, nil
}

func (j *jobManagerBase) middleware(job *work.Job, next work.NextMiddlewareFunc) error {
	start := time.Now()
	log := j.getLogForJob(job)
	log.Infof("starting job")

	jobData := models.Job{
		JobId:      job.ID,
		AccountId:  uint64(job.ArgInt64("accountId")),
		Name:       job.Name,
		Args:       job.Args,
		EnqueuedAt: time.Unix(job.EnqueuedAt, 0),
		StartedAt:  &start,
		FinishedAt: nil,
		Retries:    int(job.Fails),
	}

	if jobData.Retries == 0 {
		log.Trace("inserting job record before running")
		if _, err := j.db.Model(&jobData).Insert(&jobData); err != nil {
			log.WithError(err).Warn("failed to insert job record before running")
		}
	} else {
		log.Trace("updating job record before running")
		if _, err := j.db.Model(&jobData).WherePK().Update(&jobData); err != nil {
			log.WithError(err).Warn("failed to update job record before running")
		}
	}

	defer func() {
		log.Infof("finished job")
		if j.stats != nil {
			j.stats.JobFinished(job.Name, uint64(job.ArgInt64("accountId")), start)
		}

		now := time.Now()
		jobData.FinishedAt = &now

		log.Trace("updating job record after running")
		if _, err := j.db.Model(&jobData).WherePK().Update(&jobData); err != nil {
			log.WithError(err).Warn("failed to update job record after running")
		}
	}()

	if err := next(); err != nil {
		log.WithError(err).Errorf("failed to complete job successfully")
		return err
	}

	return nil
}

func (j *jobManagerBase) getAccountId(job *work.Job) (uint64, error) {
	accountId := job.ArgInt64("accountId")
	if accountId == 0 {
		return 0, errors.Errorf("account Id not present on job")
	}

	return uint64(accountId), nil
}

func (j *jobManagerBase) getLogForJob(job *work.Job) *logrus.Entry {
	log := j.log.WithFields(logrus.Fields{
		"job": job.Name,
		"id":  job.ID,
	})

	if accountId := job.ArgInt64("accountId"); accountId > 0 {
		log = log.WithField("accountId", accountId)
	}

	return log
}

func (j *jobManagerBase) getRepositoryForJob(job *work.Job, wrapper func(repo repository.Repository) error) error {
	accountId := uint64(job.ArgInt64("accountId"))
	if accountId == 0 {
		return errors.Errorf("account Id is required for repository access on job")
	}

	userId := uint64(job.ArgInt64("userId"))
	if userId == 0 {
		// If a user is not provided then use the system bot user.
		userId = math.MaxUint64
	}

	return j.db.RunInTransaction(context.Background(), func(txn *pg.Tx) error {
		repo := repository.NewRepositoryFromSession(userId, accountId, txn)

		return wrapper(repo)
	})
}

func (j *jobManagerBase) getJobHelperRepository(job *work.Job, wrapper func(repo repository.JobRepository) error) error {
	return j.db.RunInTransaction(context.Background(), func(txn *pg.Tx) error {
		return wrapper(repository.NewJobRepository(txn))
	})
}

func (j *jobManagerBase) Close() error {
	j.work.Stop()
	return nil
}

func (j *jobManagerBase) recover(ctx context.Context) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		if err := recover(); err != nil {
			hub.RecoverWithContext(ctx, err)
		}
	} else {
		sentry.RecoverWithContext(ctx)
	}
}
