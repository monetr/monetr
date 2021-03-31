package jobs

import (
	"github.com/gocraft/work"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
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
		fundingScheduleIds := make([]string, len(item.FundingScheduleIds))
		for x, id := range item.FundingScheduleIds {
			fundingScheduleIds[x] = strconv.FormatUint(id, 10)
		}
		_, err := j.queue.EnqueueUnique(ProcessFundingSchedules, map[string]interface{}{
			"accountId":          item.AccountId,
			"bankAccountId":      item.BankAccountId,
			"fundingScheduleIds": strings.Join(fundingScheduleIds, ","),
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
	start := time.Now()
	log := j.getLogForJob(job)
	log.Infof("processing funding schedules")

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	defer func() {
		if j.stats != nil {
			j.stats.JobFinished(PullAccountBalances, accountId, start)
		}
	}()

	bankAccountId := uint64(job.ArgInt64("bankAccountId"))
	log = log.WithField("bankAccountId", bankAccountId)

	fundingScheduleIds := make([]uint64, 0)
	idStrings := job.ArgString("fundingScheduleIds")
	for _, idString := range strings.Split(idStrings,",") {
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			log.WithError(err).Error("failed to parse funding schedule id: %s", idString)
		}

		fundingScheduleIds = append(fundingScheduleIds, id)
	}

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		account, err := repo.GetAccount()
		if err != nil {
			log.WithError(err).Error("could not retrieve account for funding schedule processing")
			return err
		}

		expensesToUpdate := make([]models.Spending, 0)

		for _, fundingScheduleId := range fundingScheduleIds {
			fundingLog := log.WithFields(logrus.Fields{
				"fundingScheduleId": fundingScheduleId,
			})

			fundingSchedule, err := repo.GetFundingSchedule(bankAccountId, fundingScheduleId)
			if err != nil {
				fundingLog.WithError(err).Error("failed to retrieve funding schedule for processing")
				return err
			}

			if time.Now().Before(fundingSchedule.NextOccurrence) {
				fundingLog.Warn("skipping processing funding schedule, it does not occur yet")
				continue
			}

			// Calculate the next time this funding schedule will happen. We need this for calculating how much each
			// expense will need the next time we do this processing.
			nextFundingOccurrence := fundingSchedule.Rule.After(time.Now(), false)
			if err = repo.UpdateNextFundingScheduleDate(fundingScheduleId, nextFundingOccurrence); err != nil {
				fundingLog.WithError(err).Error("failed to set the next occurrence for funding schedule")
				return err
			}

			// Add the funding schedule name to our logging just to make things a bit easier if we have to go look at
			// logs to find a problem.
			fundingLog = fundingLog.WithField("fundingScheduleName", fundingSchedule.Name)

			expenses, err := repo.GetExpensesByFundingSchedule(bankAccountId, fundingScheduleId)
			if err != nil {
				fundingLog.WithError(err).Error("failed to retrieve expenses for processing")
				return err
			}

			for _, expense := range expenses {
				expenseLog := fundingLog.WithFields(logrus.Fields{
					"spendingId":   expense.SpendingId,
					"spendingName": expense.Name,
				})
				if expense.TargetAmount <= expense.CurrentAmount {
					expenseLog.Trace("skipping expense, target amount is already achieved")
					continue
				}

				// TODO Take safe-to-spend into account when allocating to expenses.
				//  As of writing this I am not going to consider that balance. I'm going to assume that the user has
				//  enough money in their account at the time of this running that this will accurately reflect a real
				//  allocated balance. This can be impacted though by a delay in a deposit showing in Plaid and thus us
				//  over-allocating temporarily until the deposit shows properly in Plaid.
				expense.CurrentAmount += expense.NextContributionAmount
				if err = (&expense).CalculateNextContribution(
					account.Timezone,
					nextFundingOccurrence,
					fundingSchedule.Rule,
				); err != nil {
					expenseLog.WithError(err).Error("failed to calculate next contribution for expense")
					return err
				}

				// TODO This might cause some weird pointer behaviors.
				//  If I remember correctly using a variable that is from a "for" loop will cause issues as that
				//  variable actually changes with each iteration? So will this cause the appended value to change and
				//  thus be invalid?
				expensesToUpdate = append(expensesToUpdate, expense)
			}
		}

		log.Tracef("preparing to update %d expense(s)", len(expensesToUpdate))

		if err := repo.UpdateExpenses(bankAccountId, expensesToUpdate); err != nil {
			log.WithError(err).Error("failed to update expenses")
			return err
		}

		return nil
	})
}
