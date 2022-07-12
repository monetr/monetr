package background

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ProcessFundingSchedules = "ProcessFundingSchedules"
)

var (
	_ ScheduledJobHandler = &ProcessFundingScheduleHandler{}
)

func TriggerProcessFundingSchedules(ctx context.Context, runner JobController, args ProcessFundingScheduleArguments) error {
	return runner.triggerJob(ctx, ProcessFundingSchedules, args)
}

type ProcessFundingScheduleHandler struct {
	log          *logrus.Entry
	db           *pg.DB
	repo         repository.JobRepository
	unmarshaller JobUnmarshaller
}

func NewProcessFundingScheduleHandler(
	log *logrus.Entry,
	db *pg.DB,
) *ProcessFundingScheduleHandler {
	return &ProcessFundingScheduleHandler{
		log:          log,
		db:           db,
		repo:         repository.NewJobRepository(db),
		unmarshaller: DefaultJobUnmarshaller,
	}
}

func (p *ProcessFundingScheduleHandler) SetUnmarshaller(unmarshaller JobUnmarshaller) {
	p.unmarshaller = unmarshaller
}


func (p ProcessFundingScheduleHandler) QueueName() string {
	return ProcessFundingSchedules
}

func (p *ProcessFundingScheduleHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	job := &ProcessFundingScheduleJob{
		args: ProcessFundingScheduleArguments{},
		log:  p.log.WithContext(ctx),
		repo: nil,
	}

	if err := errors.Wrap(p.unmarshaller(data, &job.args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Process Funding Schedule job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, job.args.AccountId)

	return p.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		job.repo = repository.NewRepositoryFromSession(0, job.args.AccountId, txn)
		return job.Run(span.Context())
	})
}

func (p ProcessFundingScheduleHandler) DefaultSchedule() string {
	// Will run once an hour.
	return "0 0 * * * *"
}

func (p *ProcessFundingScheduleHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := p.log.WithContext(ctx)

	log.Info("retrieving funding schedules to process")
	fundingSchedules, err := p.repo.GetFundingSchedulesToProcess()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve funding schedules to process")
	}

	if len(fundingSchedules) == 0 {
		crumbs.Debug(ctx, "No funding schedules to be processed at this time.", nil)
		log.Info("no funding schedules to be processed at this time")
		return nil
	}

	log.WithField("count", len(fundingSchedules)).Info("preparing to enqueue funding schedules for processing")
	crumbs.Debug(ctx, "Preparing to enqueue funding schedules for processing.", map[string]interface{}{
		"count": len(fundingSchedules),
	})

	jobErrors := make([]error, 0)

	for _, item := range fundingSchedules {
		itemLog := log.WithFields(logrus.Fields{
			"accountId":          item.AccountId,
			"bankAccountId":      item.BankAccountId,
			"fundingScheduleIds": item.FundingScheduleIds,
		})
		itemLog.Trace("enqueuing funding schedules to be processed for bank account")
		err = enqueuer.EnqueueJob(ctx, p.QueueName(), ProcessFundingScheduleArguments{
			AccountId:          item.AccountId,
			BankAccountId:      item.BankAccountId,
			FundingScheduleIds: item.FundingScheduleIds,
		})
		if err != nil {
			log.WithError(err).Warn("failed to enqueue job to process funding schedule")
			crumbs.Warn(ctx, "Failed to enqueue job to process funding schedule", "job", map[string]interface{}{
				"error": err,
			})
			jobErrors = append(jobErrors, err)
			continue
		}

		itemLog.Trace("successfully enqueued funding schedules for processing")
	}

	return nil
}

type ProcessFundingScheduleArguments struct {
	AccountId          uint64   `json:"accountId"`
	BankAccountId      uint64   `json:"bankAccountId"`
	FundingScheduleIds []uint64 `json:"fundingScheduleIds"`
}

type ProcessFundingScheduleJob struct {
	args ProcessFundingScheduleArguments
	log  *logrus.Entry
	repo repository.BaseRepository
}

func (p *ProcessFundingScheduleJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec", sentry.TransactionName("Process Funding Schedules"))
	defer span.Finish()

	log := p.log.WithContext(ctx)

	account, err := p.repo.GetAccount(span.Context())
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

	initialBalances, err := p.repo.GetBalances(ctx, p.args.BankAccountId)
	if err != nil {
		log.WithError(err).Warn("failed to retrieve initial balances")
	}

	for _, fundingScheduleId := range p.args.FundingScheduleIds {
		fundingLog := log.WithFields(logrus.Fields{
			"fundingScheduleId": fundingScheduleId,
		})

		fundingSchedule, err := p.repo.GetFundingSchedule(span.Context(), p.args.BankAccountId, fundingScheduleId)
		if err != nil {
			fundingLog.WithError(err).Error("failed to retrieve funding schedule for processing")
			return err
		}

		// If this funding schedule requires waiting for a deposit to process then check to see if there are any.
		// TODO This approach is not going to scale well, if people were to create funding schedules with wait for
		//  deposit enabled. But then they never receive a deposit, or maybe the plaid link isn't active anymore, or
		//  some other scenario. We would continue to try and process these over and over again.
		if fundingSchedule.WaitForDeposit {
			log.Info("funding schedule requires a deposit to be present before processing")
			// TODO Eventually this should be moved out of the for loop.
			// TODO Maybe this could just be a count? Idk what I'd like to use these transactions for in the future.
			deposits, err := p.repo.GetRecentDepositTransactions(span.Context(), p.args.BankAccountId)
			if err != nil {
				fundingLog.WithError(err).Error("failed to retrieve recent deposits to process funding schedule")
				return err
			}

			// If there were any deposits then process the funding schedule, if there were not any deposits then do
			// nothing.
			if count := len(deposits); count > 0 {
				fundingLog.WithField("count", count).Info("found deposits in the last 24 hours")
			} else {
				fundingLog.Info("did not find any deposits in the past 24 hours, funding schedule will not be processed")
				continue
			}
		}

		if !fundingSchedule.CalculateNextOccurrence(span.Context(), timezone) {
			crumbs.Error(span.Context(), "bug: funding schedule for processing occurs in the future", "bug", map[string]interface{}{
				"nextOccurrence": fundingSchedule.NextOccurrence,
			})
			crumbs.AddTag(span.Context(), "bug", "true")
			span.Status = sentry.SpanStatusInvalidArgument
			fundingLog.Warn("skipping processing funding schedule, it does not occur yet")
			continue
		}

		if err = p.repo.UpdateNextFundingScheduleDate(span.Context(), fundingScheduleId, fundingSchedule.NextOccurrence); err != nil {
			fundingLog.WithError(err).Error("failed to set the next occurrence for funding schedule")
			return err
		}

		expenses, err := p.repo.GetSpendingByFundingSchedule(span.Context(), p.args.BankAccountId, fundingScheduleId)
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
			for i := range expenses {
				// :shakes-fist: arrays
				spending := expenses[i]
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
					fundingSchedule,
					time.Now(),
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

	if err = p.repo.UpdateSpending(span.Context(), p.args.BankAccountId, expensesToUpdate); err != nil {
		log.WithError(err).Error("failed to update spending")
		return err
	}

	updatedBalances, err := p.repo.GetBalances(ctx, p.args.BankAccountId)
	if err != nil {
		log.WithError(err).Warn("failed to retrieve updated balances")
	}

	// Trying to determine how often balances go negative.
	crumbs.Debug(ctx, "Funding result balances", map[string]interface{}{
		"before": initialBalances,
		"after":  updatedBalances,
	})
	if initialBalances.Safe > 0 && updatedBalances.Safe < 0 {
		crumbs.Warn(ctx, "Safe to spend has gone negative!", "balance", nil)
		crumbs.AddTag(ctx, "safe-to-spend", "negative")
	} else if updatedBalances.Safe > 0 {
		crumbs.AddTag(ctx, "safe-to-spend", "positive")
	}

	return nil
}
