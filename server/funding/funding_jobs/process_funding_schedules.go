package funding_jobs

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type ProcessFundingScheduleArguments struct {
	AccountId          models.ID[models.Account]           `json:"accountId"`
	BankAccountId      models.ID[models.BankAccount]       `json:"bankAccountId"`
	FundingScheduleIds []models.ID[models.FundingSchedule] `json:"fundingScheduleIds"`
}

func ProcessFundingSchedulesCron(ctx queue.Context) error {
	log := ctx.Log()

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())

	log.InfoContext(ctx, "retrieving funding schedules to process")
	fundingSchedules, err := jobRepo.GetFundingSchedulesToProcess(ctx)
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

	for _, item := range fundingSchedules {
		itemLog := log.With(
			"accountId", item.AccountId,
			"bankAccountId", item.BankAccountId,
			"fundingScheduleIds", item.FundingScheduleIds,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing funding schedules to be processed for bank account")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			ProcessFundingSchedule,
			ProcessFundingScheduleArguments{
				AccountId:          item.AccountId,
				BankAccountId:      item.BankAccountId,
				FundingScheduleIds: item.FundingScheduleIds,
			},
		); err != nil {
			log.WarnContext(
				ctx,
				"failed to enqueue job to process funding schedule",
				"err", err,
			)
			crumbs.Warn(
				ctx,
				"Failed to enqueue job to process funding schedule",
				"job",
				map[string]any{
					"error": err,
				},
			)
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued funding schedules for processing")
	}

	return nil
}

func ProcessFundingSchedule(ctx queue.Context, args ProcessFundingScheduleArguments) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		span := sentry.SpanFromContext(ctx)
		log := ctx.Log().With("bankAccountId", args.BankAccountId)

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_system",
			args.AccountId,
			ctx.DB(),
			log,
		)

		account, err := repo.GetAccount(ctx)
		if err != nil {
			log.ErrorContext(ctx, "could not retrieve account for funding schedule processing", "err", err)
			return err
		}

		timezone, err := account.GetTimezone()
		if err != nil {
			log.ErrorContext(ctx, "could not parse account's timezone", "err", err)
			return err
		}

		expensesToUpdate := make([]models.Spending, 0)

		initialBalances, err := repo.GetBalances(ctx, args.BankAccountId)
		if err != nil {
			log.WarnContext(ctx, "failed to retrieve initial balances", "err", err)
		}

		for _, fundingScheduleId := range args.FundingScheduleIds {
			fundingLog := log.With("fundingScheduleId", fundingScheduleId)

			fundingSchedule, err := repo.GetFundingSchedule(ctx, args.BankAccountId, fundingScheduleId)
			if err != nil {
				fundingLog.ErrorContext(ctx, "failed to retrieve funding schedule for processing", "err", err)
				return err
			}

			// If this funding schedule requires waiting for a deposit to process then
			// check to see if there are any.
			// TODO This approach is not going to scale well, if people were to create
			// funding schedules with wait for deposit enabled. But then they never
			// receive a deposit, or maybe the plaid link isn't active anymore, or some
			// other scenario. We would continue to try and process these over and over
			// again.
			if fundingSchedule.WaitForDeposit {
				log.InfoContext(ctx, "funding schedule requires a deposit to be present before processing")
				// TODO Eventually this should be moved out of the for loop.
				// TODO Maybe this could just be a count? Idk what I'd like to use these
				// transactions for in the future.
				deposits, err := repo.GetRecentDepositTransactions(ctx, args.BankAccountId)
				if err != nil {
					fundingLog.ErrorContext(ctx, "failed to retrieve recent deposits to process funding schedule", "err", err)
					return err
				}

				// If there were any deposits then process the funding schedule, if there
				// were not any deposits then do nothing.
				if count := len(deposits); count > 0 {
					fundingLog.InfoContext(ctx, fmt.Sprintf("found %d deposits in the last 24 hours", count))
				} else {
					fundingLog.InfoContext(ctx, "did not find any deposits in the past 24 hours, funding schedule will not be processed")
					continue
				}
			}

			if !fundingSchedule.CalculateNextOccurrence(ctx, ctx.Clock().Now(), timezone) {
				crumbs.IndicateBug(ctx, "bug: funding schedule for processing occurs in the future", map[string]any{
					"nextOccurrence": fundingSchedule.NextRecurrence,
				})
				if span != nil {
					span.Status = sentry.SpanStatusInvalidArgument
				}
				fundingLog.WarnContext(ctx, "skipping processing funding schedule, it does not occur yet")
				continue
			}

			if err = repo.UpdateFundingSchedule(ctx, fundingSchedule); err != nil {
				fundingLog.ErrorContext(ctx, "failed to update the funding schedule with the updated next recurrence", "err", err)
				return err
			}

			expenses, err := repo.GetSpendingByFundingSchedule(ctx, args.BankAccountId, fundingScheduleId)
			if err != nil {
				fundingLog.ErrorContext(ctx, "failed to retrieve expenses for processing", "err", err)
				return err
			}

			switch len(expenses) {
			case 0:
				crumbs.Debug(
					ctx,
					"There are no spending objects associated with this funding schedule",
					map[string]any{
						"fundingScheduleId": fundingScheduleId,
					},
				)
			default:
				for i := range expenses {
					spending := expenses[i]
					spendingLog := fundingLog.With("spendingId", spending.SpendingId)

					if spending.IsPaused {
						crumbs.Debug(
							ctx,
							"Spending object is paused, it will be skipped",
							map[string]any{
								"fundingScheduleId": fundingScheduleId,
								"spendingId":        spending.SpendingId,
							},
						)
						spendingLog.DebugContext(ctx, "skipping funding spending item, it is paused")
						continue
					}

					progressAmount := spending.GetProgressAmount()

					if spending.TargetAmount <= progressAmount {
						crumbs.Debug(
							ctx,
							"Spending object already has target amount, it will be skipped",
							map[string]any{
								"fundingScheduleId": fundingScheduleId,
								"spendingId":        spending.SpendingId,
							},
						)
						spendingLog.Log(
							ctx,
							logging.LevelTrace,
							"skipping spending, target amount is already achieved",
						)
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
						ctx,
						timezone,
						fundingSchedule,
						ctx.Clock().Now(),
						log,
					)

					expensesToUpdate = append(expensesToUpdate, spending)
				}
			}
		}

		if len(expensesToUpdate) == 0 {
			crumbs.Debug(ctx, "No spending objects to update for funding schedule", nil)
			log.InfoContext(ctx, "no spending objects to update for funding schedule")
			return nil
		}

		log.DebugContext(ctx, fmt.Sprintf("preparing to update %d spending(s)", len(expensesToUpdate)))

		crumbs.Debug(ctx, "Updating spending objects with recalculated contributions", map[string]any{
			"count": len(expensesToUpdate),
		})

		if err = repo.UpdateSpending(ctx, args.BankAccountId, expensesToUpdate); err != nil {
			log.ErrorContext(ctx, "failed to update spending", "err", err)
			return err
		}

		updatedBalances, err := repo.GetBalances(ctx, args.BankAccountId)
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
	})
}
