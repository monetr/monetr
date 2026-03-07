package background

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

const (
	ProcessFundingSchedules = "ProcessFundingSchedules"
)

var (
	_ ScheduledJobHandler = &ProcessFundingScheduleHandler{}
)

func TriggerProcessFundingSchedules(ctx context.Context, runner JobController, args ProcessFundingScheduleArguments) error {
	return runner.EnqueueJob(ctx, ProcessFundingSchedules, args)
}

type ProcessFundingScheduleHandler struct {
	log          *slog.Logger
	db           *pg.DB
	repo         repository.JobRepository
	unmarshaller JobUnmarshaller
	clock        clock.Clock
}

func NewProcessFundingScheduleHandler(
	log *slog.Logger,
	db *pg.DB,
	clock clock.Clock,
) *ProcessFundingScheduleHandler {
	return &ProcessFundingScheduleHandler{
		log:          log,
		db:           db,
		repo:         repository.NewJobRepository(db, clock),
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (p ProcessFundingScheduleHandler) QueueName() string {
	return ProcessFundingSchedules
}

func (p *ProcessFundingScheduleHandler) HandleConsumeJob(
	ctx context.Context,
	log *slog.Logger,
	data []byte,
) error {
	var args ProcessFundingScheduleArguments
	if err := errors.Wrap(p.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Process Funding Schedule job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return p.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()
		log = log.With(
			"accountId", args.AccountId,
			"bankAccountId", args.BankAccountId,
		)
		job := &ProcessFundingScheduleJob{
			args:  args,
			log:   log,
			repo:  nil,
			clock: p.clock,
		}

		job.repo = repository.NewRepositoryFromSession(
			p.clock,
			"user_system",
			job.args.AccountId,
			txn,
			log,
		)
		return job.Run(span.Context())
	})
}

func (p ProcessFundingScheduleHandler) DefaultSchedule() string {
	// Will run once an hour.
	return "0 0 * * * *"
}

func (p *ProcessFundingScheduleHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := p.log

	log.InfoContext(ctx, "retrieving funding schedules to process")
	fundingSchedules, err := p.repo.GetFundingSchedulesToProcess(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve funding schedules to process")
	}

	if len(fundingSchedules) == 0 {
		crumbs.Debug(ctx, "No funding schedules to be processed at this time.", nil)
		log.InfoContext(ctx, "no funding schedules to be processed at this time")
		return nil
	}

	log.InfoContext(ctx, "preparing to enqueue funding schedules for processing", "count", len(fundingSchedules))
	crumbs.Debug(ctx, "Preparing to enqueue funding schedules for processing.", map[string]any{
		"count": len(fundingSchedules),
	})

	jobErrors := make([]error, 0)

	for _, item := range fundingSchedules {
		itemLog := log.With(
			"accountId", item.AccountId,
			"bankAccountId", item.BankAccountId,
			"fundingScheduleIds", item.FundingScheduleIds,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing funding schedules to be processed for bank account")
		err = enqueuer.EnqueueJob(ctx, p.QueueName(), ProcessFundingScheduleArguments{
			AccountId:          item.AccountId,
			BankAccountId:      item.BankAccountId,
			FundingScheduleIds: item.FundingScheduleIds,
		})
		if err != nil {
			log.WarnContext(ctx, "failed to enqueue job to process funding schedule", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to process funding schedule", "job", map[string]any{
				"error": err,
			})
			jobErrors = append(jobErrors, err)
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued funding schedules for processing")
	}

	return nil
}

type ProcessFundingScheduleArguments struct {
	AccountId          ID[Account]           `json:"accountId"`
	BankAccountId      ID[BankAccount]       `json:"bankAccountId"`
	FundingScheduleIds []ID[FundingSchedule] `json:"fundingScheduleIds"`
}

type ProcessFundingScheduleJob struct {
	args  ProcessFundingScheduleArguments
	log   *slog.Logger
	repo  repository.BaseRepository
	clock clock.Clock
}

func (p *ProcessFundingScheduleJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := p.log.With("bankAccountId", p.args.BankAccountId)

	account, err := p.repo.GetAccount(span.Context())
	if err != nil {
		log.ErrorContext(span.Context(), "could not retrieve account for funding schedule processing", "err", err)
		return err
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		log.ErrorContext(span.Context(), "could not parse account's timezone", "err", err)
		return err
	}

	expensesToUpdate := make([]Spending, 0)

	initialBalances, err := p.repo.GetBalances(ctx, p.args.BankAccountId)
	if err != nil {
		log.WarnContext(ctx, "failed to retrieve initial balances", "err", err)
	}

	for _, fundingScheduleId := range p.args.FundingScheduleIds {
		fundingLog := log.With("fundingScheduleId", fundingScheduleId)

		fundingSchedule, err := p.repo.GetFundingSchedule(span.Context(), p.args.BankAccountId, fundingScheduleId)
		if err != nil {
			fundingLog.ErrorContext(span.Context(), "failed to retrieve funding schedule for processing", "err", err)
			return err
		}

		// If this funding schedule requires waiting for a deposit to process then check to see if there are any.
		// TODO This approach is not going to scale well, if people were to create funding schedules with wait for
		//  deposit enabled. But then they never receive a deposit, or maybe the plaid link isn't active anymore, or
		//  some other scenario. We would continue to try and process these over and over again.
		if fundingSchedule.WaitForDeposit {
			log.InfoContext(span.Context(), "funding schedule requires a deposit to be present before processing")
			// TODO Eventually this should be moved out of the for loop.
			// TODO Maybe this could just be a count? Idk what I'd like to use these transactions for in the future.
			deposits, err := p.repo.GetRecentDepositTransactions(span.Context(), p.args.BankAccountId)
			if err != nil {
				fundingLog.ErrorContext(span.Context(), "failed to retrieve recent deposits to process funding schedule", "err", err)
				return err
			}

			// If there were any deposits then process the funding schedule, if there were not any deposits then do
			// nothing.
			if count := len(deposits); count > 0 {
				fundingLog.InfoContext(span.Context(), fmt.Sprintf("found %d deposits in the last 24 hours", count))
			} else {
				fundingLog.InfoContext(span.Context(), "did not find any deposits in the past 24 hours, funding schedule will not be processed")
				continue
			}
		}

		if !fundingSchedule.CalculateNextOccurrence(span.Context(), p.clock.Now(), timezone) {
			crumbs.IndicateBug(span.Context(), "bug: funding schedule for processing occurs in the future", map[string]any{
				"nextOccurrence": fundingSchedule.NextRecurrence,
			})
			span.Status = sentry.SpanStatusInvalidArgument
			fundingLog.WarnContext(span.Context(), "skipping processing funding schedule, it does not occur yet")
			continue
		}

		if err = p.repo.UpdateFundingSchedule(span.Context(), fundingSchedule); err != nil {
			fundingLog.ErrorContext(span.Context(), "failed to update the funding schedule with the updated next recurrence", "err", err)
			return err
		}

		expenses, err := p.repo.GetSpendingByFundingSchedule(span.Context(), p.args.BankAccountId, fundingScheduleId)
		if err != nil {
			fundingLog.ErrorContext(span.Context(), "failed to retrieve expenses for processing", "err", err)
			return err
		}

		switch len(expenses) {
		case 0:
			crumbs.Debug(span.Context(), "There are no spending objects associated with this funding schedule", map[string]any{
				"fundingScheduleId": fundingScheduleId,
			})
		default:
			for i := range expenses {
				spending := expenses[i]
				spendingLog := fundingLog.With("spendingId", spending.SpendingId)

				if spending.IsPaused {
					crumbs.Debug(span.Context(), "Spending object is paused, it will be skipped", map[string]any{
						"fundingScheduleId": fundingScheduleId,
						"spendingId":        spending.SpendingId,
					})
					spendingLog.DebugContext(span.Context(), "skipping funding spending item, it is paused")
					continue
				}

				progressAmount := spending.GetProgressAmount()

				if spending.TargetAmount <= progressAmount {
					crumbs.Debug(span.Context(), "Spending object already has target amount, it will be skipped", map[string]any{
						"fundingScheduleId": fundingScheduleId,
						"spendingId":        spending.SpendingId,
					})
					spendingLog.Log(span.Context(), logging.LevelTrace, "skipping spending, target amount is already achieved")
					continue
				}

				// TODO Take free-to-use into account when allocating to expenses.
				//  As of writing this I am not going to consider that balance. I'm
				//  going to assume that the user has enough money in their account at
				//  the time of this running that this will accurately reflect a real
				//  allocated balance. This can be impacted though by a delay in a
				//  deposit showing in Plaid and thus us over-allocating temporarily
				//  until the deposit shows properly in Plaid.
				spending.CurrentAmount += spending.NextContributionAmount
				(&spending).CalculateNextContribution(
					span.Context(),
					timezone,
					fundingSchedule,
					p.clock.Now(),
					log,
				)

				expensesToUpdate = append(expensesToUpdate, spending)
			}
		}
	}

	if len(expensesToUpdate) == 0 {
		crumbs.Debug(span.Context(), "No spending objects to update for funding schedule", nil)
		log.InfoContext(span.Context(), "no spending objects to update for funding schedule")
		return nil
	}

	log.DebugContext(span.Context(), fmt.Sprintf("preparing to update %d spending(s)", len(expensesToUpdate)))

	crumbs.Debug(span.Context(), "Updating spending objects with recalculated contributions", map[string]any{
		"count": len(expensesToUpdate),
	})

	if err = p.repo.UpdateSpending(span.Context(), p.args.BankAccountId, expensesToUpdate); err != nil {
		log.ErrorContext(span.Context(), "failed to update spending", "err", err)
		return err
	}

	updatedBalances, err := p.repo.GetBalances(ctx, p.args.BankAccountId)
	if err != nil {
		log.WarnContext(ctx, "failed to retrieve updated balances", "err", err)
	}

	// Trying to determine how often balances go negative.
	crumbs.Debug(ctx, "Funding result balances", map[string]any{
		"before": initialBalances,
		"after":  updatedBalances,
	})
	if initialBalances.Free > 0 && updatedBalances.Free < 0 {
		crumbs.Warn(ctx, "Free to use has gone negative!", "balance", nil)
		crumbs.AddTag(ctx, "free-to-use", "negative")
	} else if updatedBalances.Free > 0 {
		crumbs.AddTag(ctx, "free-to-use", "positive")
	}

	return nil
}
