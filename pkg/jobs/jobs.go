package jobs

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/metrics"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

type JobManager interface {
	TriggerPullInitialTransactions(accountId, userId, linkId uint64) (jobId string, err error)
	Close() error
}

type jobManagerBase struct {
	log         *logrus.Entry
	work        *work.WorkerPool
	queue       *work.Enqueuer
	db          *pg.DB
	plaidClient *plaid.Client
	stats       *metrics.Stats
}

func NewJobManager(log *logrus.Entry, pool *redis.Pool, db *pg.DB, plaidClient *plaid.Client, stats *metrics.Stats) JobManager {
	manager := &jobManagerBase{
		log: log,
		// TODO (elliotcourant) Use namespace from config.
		work:        work.NewWorkerPool(struct{}{}, 4, "harder", pool),
		queue:       work.NewEnqueuer("harder", pool),
		db:          db,
		plaidClient: plaidClient,
		stats:       stats,
	}

	manager.work.Middleware(manager.middleware)

	manager.work.Job(EnqueueCheckPendingTransactions, manager.enqueueCheckPendingTransactions)
	manager.work.Job(EnqueueProcessFundingSchedules, manager.enqueueProcessFundingSchedules)
	manager.work.Job(EnqueuePullAccountBalances, manager.enqueuePullAccountBalances)
	manager.work.Job(EnqueuePullLatestTransactions, manager.enqueuePullLatestTransactions)

	manager.work.Job(CheckPendingTransactions, manager.checkPendingTransactions)
	manager.work.Job(ProcessFundingSchedules, manager.processFundingSchedules)
	manager.work.Job(PullAccountBalances, manager.pullAccountBalances)
	manager.work.Job(PullInitialTransactions, manager.pullInitialTransactions)
	manager.work.Job(PullLatestTransactions, manager.pullLatestTransactions)

	// Every 30 minutes. 0 */30 * * * *

	// Every hour.
	manager.work.PeriodicallyEnqueue("0 0 * * * *", EnqueuePullAccountBalances)
	manager.work.PeriodicallyEnqueue("0 0 * * * *", EnqueuePullLatestTransactions)
	manager.work.PeriodicallyEnqueue("0 0 * * * *", EnqueueProcessFundingSchedules)
	manager.work.PeriodicallyEnqueue("0 0 * * * *", EnqueueCheckPendingTransactions)

	manager.work.Start()

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
	log := j.log.WithFields(logrus.Fields{
		"jobId": job.ID,
		"name":  job.Name,
	})
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
	return next()
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
