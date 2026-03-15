package spending_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type ProcessSpendingArguments struct {
	AccountId     models.ID[models.Account]     `json:"accountId"`
	BankAccountId models.ID[models.BankAccount] `json:"bankAccountId"`
}

func ProcessSpendingCron(ctx queue.Context) error {
	log := ctx.Log()

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())
	bankAccountsWithStaleSpending, err := jobRepo.GetBankAccountsWithStaleSpending(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve bank accounts with stale spending")
	}

	if len(bankAccountsWithStaleSpending) == 0 {
		crumbs.Debug(ctx, "No bank accounts had stale spending objects.", nil)
		log.InfoContext(ctx, "no bank accounts have stale spending objects")
		return nil
	}

	log.InfoContext(ctx, "found bank accounts with stale spending", "count", len(bankAccountsWithStaleSpending))
	crumbs.Debug(ctx, "Found bank accounts with stale spending.", map[string]any{
		"count": len(bankAccountsWithStaleSpending),
	})

	jobErrors := make([]error, 0)

	for _, item := range bankAccountsWithStaleSpending {
		itemLog := log.With(
			"accountId", item.AccountId,
			"bankAccountId", item.BankAccountId,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing bank account to process stale spending")

		err = queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			ProcessSpending,
			ProcessSpendingArguments{
				AccountId:     item.AccountId,
				BankAccountId: item.BankAccountId,
			},
		)
		if err != nil {
			log.WarnContext(ctx, "failed to enqueue job to process stale spending", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to process stale spending", "job", map[string]any{
				"error": err,
			})
			jobErrors = append(jobErrors, err)
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued bank accounts for stale spending processing")
	}

	return nil
}

func ProcessSpending(ctx queue.Context, args ProcessSpendingArguments) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		crumbs.IncludeUserInScope(ctx, args.AccountId)
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"bankAccountId", args.BankAccountId,
		)

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_system",
			args.AccountId,
			ctx.DB(),
			log,
		)

		account, err := repo.GetAccount(ctx)
		if err != nil {
			log.ErrorContext(ctx, "failed to retrieve account to process stale spending", "err", err)
			return err
		}

		timezone, err := account.GetTimezone()
		if err != nil {
			log.ErrorContext(ctx, "failed to parse account timezone", "err", err)
			return err
		}

		now := ctx.Clock().Now()
		allSpending, err := repo.GetSpending(ctx, args.BankAccountId)
		if err != nil {
			log.ErrorContext(ctx, "failed to retrieve spending for bank account", "err", err)
			return err
		}

		fundingSchedules := map[models.ID[models.FundingSchedule]]*models.FundingSchedule{}

		spendingToUpdate := make([]models.Spending, 0, len(allSpending))
		for i := range allSpending {
			// Avoid funky pointer issues with arrays and for loops.
			spending := allSpending[i]

			// Skip spending objects that are not stale, or ones that are paused.
			if !spending.GetIsStale(now) || spending.GetIsPaused() {
				continue
			}

			fundingSchedule, ok := fundingSchedules[spending.FundingScheduleId]
			if !ok {
				fundingSchedule, err = repo.GetFundingSchedule(
					ctx,
					spending.BankAccountId,
					spending.FundingScheduleId,
				)
				if err != nil {
					log.WarnContext(ctx, "failed to retrieve funding schedule for spending object, it will not be processed", "err", err)
					continue
				}

				fundingSchedules[spending.FundingScheduleId] = fundingSchedule
			}

			spending.CalculateNextContribution(
				ctx,
				timezone,
				fundingSchedule,
				now,
				log,
			)

			spendingToUpdate = append(spendingToUpdate, spending)
		}

		if len(spendingToUpdate) == 0 {
			log.InfoContext(ctx, "no stale spending object were updated")
			return nil
		}

		log.InfoContext(ctx, "updating stale spending objects", "count", len(spendingToUpdate))

		return errors.Wrap(repo.UpdateSpending(
			ctx,
			args.BankAccountId,
			spendingToUpdate,
		), "failed to update stale spending")
	})
}
