package jobs

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/monetr/rest-api/pkg/util"
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
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Process Funding Schedules"))
	defer span.Finish()

	log := j.getLogForJob(job)
	log.Infof("processing funding schedules")

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	span.SetTag("accountId", strconv.FormatUint(accountId, 10))

	bankAccountId := uint64(job.ArgInt64("bankAccountId"))
	log = log.WithField("bankAccountId", bankAccountId)
	span.SetTag("bankAccountId", strconv.FormatUint(bankAccountId, 10))

	fundingScheduleIds := make([]uint64, 0)
	idStrings := job.ArgString("fundingScheduleIds")
	log = log.WithField("fundingScheduleIds", idStrings)
	for _, idString := range strings.Split(idStrings, ",") {
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			log.WithError(err).Errorf("failed to parse funding schedule id: %s", idString)
		}

		fundingScheduleIds = append(fundingScheduleIds, id)
	}

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		account, err := repo.GetAccount(span.Context())
		if err != nil {
			log.WithError(err).Error("could not retrieve account for funding schedule processing")
			return err
		}

		timezone, err := account.GetTimezone()
		if err != nil {
			log.WithError(err).Error("could not parse account's timezone")
			return err
		}

		expensesToUpdate := make([]models.Spending, 0)

		for _, fundingScheduleId := range fundingScheduleIds {
			fundingLog := log.WithFields(logrus.Fields{
				"fundingScheduleId": fundingScheduleId,
			})

			fundingSchedule, err := repo.GetFundingSchedule(span.Context(), bankAccountId, fundingScheduleId)
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
			nextFundingOccurrence := util.MidnightInLocal(fundingSchedule.Rule.After(time.Now(), false), timezone)
			if err = repo.UpdateNextFundingScheduleDate(span.Context(), fundingScheduleId, nextFundingOccurrence); err != nil {
				fundingLog.WithError(err).Error("failed to set the next occurrence for funding schedule")
				return err
			}

			// Add the funding schedule name to our logging just to make things a bit easier if we have to go look at
			// logs to find a problem.
			fundingLog = fundingLog.WithField("fundingScheduleName", fundingSchedule.Name)

			expenses, err := repo.GetSpendingByFundingSchedule(span.Context(), bankAccountId, fundingScheduleId)
			if err != nil {
				fundingLog.WithError(err).Error("failed to retrieve expenses for processing")
				return err
			}

			for _, spending := range expenses {
				spendingLog := fundingLog.WithFields(logrus.Fields{
					"spendingId":   spending.SpendingId,
					"spendingName": spending.Name,
				})

				if spending.IsPaused {
					spendingLog.Debug("skipping funding spending item, it is paused")
					continue
				}

				progressAmount := spending.GetProgressAmount()

				if spending.TargetAmount <= progressAmount {
					spendingLog.Trace("skipping spending, target amount is already achieved")
					continue
				}

				// TODO Take safe-to-spend into account when allocating to expenses.
				//  As of writing this I am not going to consider that balance. I'm going to assume that the user has
				//  enough money in their account at the time of this running that this will accurately reflect a real
				//  allocated balance. This can be impacted though by a delay in a deposit showing in Plaid and thus us
				//  over-allocating temporarily until the deposit shows properly in Plaid.
				spending.CurrentAmount += spending.NextContributionAmount
				if err = (&spending).CalculateNextContribution(
					span.Context(),
					account.Timezone,
					nextFundingOccurrence,
					fundingSchedule.Rule,
				); err != nil {
					spendingLog.WithError(err).Error("failed to calculate next contribution for spending")
					return err
				}

				// TODO This might cause some weird pointer behaviors.
				//  If I remember correctly using a variable that is from a "for" loop will cause issues as that
				//  variable actually changes with each iteration? So will this cause the appended value to change and
				//  thus be invalid?
				expensesToUpdate = append(expensesToUpdate, spending)
			}
		}

		if len(expensesToUpdate) == 0 {
			log.Info("no spending objects to update for funding schedule")
			return nil
		}

		log.Debugf("preparing to update %d spending(s)", len(expensesToUpdate))

		if err := repo.UpdateSpending(span.Context(), bankAccountId, expensesToUpdate); err != nil {
			log.WithError(err).Error("failed to update spending")
			return err
		}

		return nil
	})
}
