package jobs

import (
	"github.com/gocraft/work"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/pkg/errors"
)

const (
	EnqueueProcessFundingSchedules = "EnqueueProcessFundingSchedules"
	ProcessFundingSchedules        = "ProcessFundingSchedules"
)

func (j *jobManagerBase) enqueueProcessFundingSchedules(job *work.Job) error {
	log := j.getLogForJob(job)

	var items []repository.ProcessFundingSchedulesItem
	err := j.getJobHelperRepository(job, func(repo repository.JobRepository) (err error) {
		items, err = repo.GetFundingSchedulesToProcess()
		return err
	})
	if err != nil {
		// TODO (elliotcourant) Related to the todo in the repo method, if the error is a no rows error what do we want
		//  to do here?
		return err
	}

	if len(items) == 0 {
		log.Info("no funding schedules to process")
		return nil
	}

	log.Infof("enqueueing %d funding schedule(s) to be processed", len(items))

	for _, item := range items {
		accountLog := log.WithField("accountId", item.AccountId)
		accountLog.Trace("enqueueing for funding schedule processing")
		_, err := j.queue.EnqueueUnique(ProcessFundingSchedules, map[string]interface{}{
			"accountId":          item.AccountId,
			"fundingScheduleIds": item.FundingScheduleIds,
		})
		if err != nil {
			err = errors.Wrap(err, "failed to enqueue account")
			accountLog.WithError(err).Error("could not enqueue account, funding schedules will not be processed")
			continue
		}
		accountLog.Trace("successfully enqueued account for funding schedule processing")
	}

	return nil

}

func (j *jobManagerBase) processFundingSchedules(job *work.Job) error {
	return nil
}
