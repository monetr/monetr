package jobs

import (
	"github.com/go-pg/pg/v10"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type JobManager struct {
	log   *logrus.Entry
	work  *work.WorkerPool
	queue *work.Enqueuer
	db    *pg.DB
}

func NewJobManager(log *logrus.Entry, pool *redis.Pool, db *pg.DB) *JobManager {
	manager := &JobManager{
		log: log,
		// TODO (elliotcourant) Use namespace from config.
		work:  work.NewWorkerPool(struct{}{}, 4, "harder", pool),
		queue: work.NewEnqueuer("harder", pool),
		db:    nil,
	}

	manager.work.Job(EnqueuePullAccountBalances, manager.EnqueuePullAccountBalances)
	manager.work.Job(PullAccountBalances, manager.PullAccountBalances)

	manager.work.PeriodicallyEnqueue("*/30 * * * *", EnqueuePullAccountBalances)

	return manager
}

func (j *JobManager) getAccountId(job *work.Job) (uint64, error) {
	accountId, ok := job.Args["accountId"]
	if !ok {
		return 0, errors.Errorf("account Id not present on job")
	}

	// TODO (elliotcourant) This won't work, it probably wont unmarshal it back into a uint64.
	return accountId.(uint64), nil
}

func (j *JobManager) Drain() error {
	j.work.Drain()
	return nil
}
