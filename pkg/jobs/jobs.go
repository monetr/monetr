package jobs

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"math"
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
}

func NewJobManager(log *logrus.Entry, pool *redis.Pool, db *pg.DB, plaidClient *plaid.Client) JobManager {
	manager := &jobManagerBase{
		log: log,
		// TODO (elliotcourant) Use namespace from config.
		work:        work.NewWorkerPool(struct{}{}, 4, "harder", pool),
		queue:       work.NewEnqueuer("harder", pool),
		db:          db,
		plaidClient: plaidClient,
	}

	manager.work.Middleware(manager.middleware)

	manager.work.Job(EnqueueProcessFundingSchedules, manager.enqueueProcessFundingSchedules)
	manager.work.Job(EnqueuePullAccountBalances, manager.enqueuePullAccountBalances)
	manager.work.Job(EnqueuePullLatestTransactions, manager.enqueuePullLatestTransactions)

	manager.work.Job(ProcessFundingSchedules, manager.processFundingSchedules)
	manager.work.Job(PullAccountBalances, manager.pullAccountBalances)
	manager.work.Job(PullInitialTransactions, manager.pullInitialTransactions)
	manager.work.Job(PullLatestTransactions, manager.pullLatestTransactions)

	manager.work.PeriodicallyEnqueue("0 */30 * * * *", EnqueuePullAccountBalances)
	manager.work.PeriodicallyEnqueue("0 */30 * * * *", EnqueuePullLatestTransactions)
	manager.work.PeriodicallyEnqueue("0 0 * * * *", EnqueueProcessFundingSchedules)

	manager.work.Start()

	return manager
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
	j.log.WithField("jobId", job.ID).WithField("name", job.Name).Infof("starting job")
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
