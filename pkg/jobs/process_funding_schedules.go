package jobs

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetr/rest-api/pkg/crumbs"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
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

func (j *jobManagerBase) processFundingSchedules(job *work.Job) (err error) {
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Process Funding Schedules"))
	defer span.Finish()

	log := j.getLogForJob(job)
	log.Infof("processing funding schedules")

	defer func() {
		if err != nil {
			hub.CaptureException(err)
		}
	}()

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	bankAccountId := uint64(job.ArgInt64("bankAccountId"))
	log = log.WithField("bankAccountId", bankAccountId)

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

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       strconv.FormatUint(accountId, 10),
			Username: fmt.Sprintf("account:%d", accountId),
		})
		scope.SetTag("accountId", strconv.FormatUint(accountId, 10))
		scope.SetTag("bankAccountId", strconv.FormatUint(bankAccountId, 10))
		scope.SetTag("jobId", job.ID)
	})

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

			if !fundingSchedule.CalculateNextOccurrence(span.Context(), timezone) {
				fundingLog.Warn("skipping processing funding schedule, it does not occur yet")
				continue
			}

			if err = repo.UpdateNextFundingScheduleDate(span.Context(), fundingScheduleId, fundingSchedule.NextOccurrence); err != nil {
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

			switch len(expenses) {
			case 0:
				crumbs.Debug(span.Context(), "There are no spending objects associated with this funding schedule", map[string]interface{}{
					"fundingScheduleId": fundingScheduleId,
				})
			default:
				for _, spending := range expenses {
					spendingLog := fundingLog.WithFields(logrus.Fields{
						"spendingId":   spending.SpendingId,
						"spendingName": spending.Name,
					})

					if spending.IsPaused {
						crumbs.Debug(span.Context(), "Spending object is paused, it will be skipped", map[string]interface{}{
							"fundingScheduleId": fundingScheduleId,
							"spendingId":        spending.SpendingId,
						})
						spendingLog.Debug("skipping funding spending item, it is paused")
						continue
					}

					progressAmount := spending.GetProgressAmount()

					if spending.TargetAmount <= progressAmount {
						crumbs.Debug(span.Context(), "Spending object already has target amount, it will be skipped", map[string]interface{}{
							"fundingScheduleId": fundingScheduleId,
							"spendingId":        spending.SpendingId,
						})
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
						fundingSchedule.NextOccurrence,
						fundingSchedule.Rule,
					); err != nil {
						crumbs.Error(span.Context(), "Failed to calculate next contribution for spending", "spending", map[string]interface{}{
							"fundingScheduleId": fundingScheduleId,
							"spendingId":        spending.SpendingId,
						})
						spendingLog.WithError(err).Error("failed to calculate next contribution for spending")
						return err
					}

					expensesToUpdate = append(expensesToUpdate, spending)
				}
			}

		}

		if len(expensesToUpdate) == 0 {
			crumbs.Debug(span.Context(), "No spending objects to update for funding schedule", nil)
			log.Info("no spending objects to update for funding schedule")
			return nil
		}

		log.Debugf("preparing to update %d spending(s)", len(expensesToUpdate))

		crumbs.Debug(span.Context(), "Updating spending objects with recalculated contributions", map[string]interface{}{
			"count": len(expensesToUpdate),
		})

		if err = repo.UpdateSpending(span.Context(), bankAccountId, expensesToUpdate); err != nil {
			log.WithError(err).Error("failed to update spending")
			return err
		}

		return nil
	})
}
