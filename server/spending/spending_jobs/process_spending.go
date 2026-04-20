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

		// Only manual links can have transactions auto-created on their behalf;
		// Plaid links already receive real transactions from the institution.
		isManual, err := repo.GetLinkIsManualByBankAccountId(ctx, args.BankAccountId)
		if err != nil {
			log.ErrorContext(ctx, "failed to determine if link is manual", "err", err)
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
		transactionsToCreate := make([]models.Transaction, 0)
		var bankAccount *models.BankAccount
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

			// Capture the due date before CalculateNextContribution bumps the
			// recurrence forward; this is the date the expense was due.
			dueDate := spending.NextRecurrence
			spending.CalculateNextContribution(
				ctx,
				timezone,
				fundingSchedule,
				now,
				log,
			)

			// If this expense has auto create transaction enabled, create a
			// transaction for the just-passed due date and allocate it from the
			// expense. Goals are excluded by design.
			if isManual &&
				spending.SpendingType == models.SpendingTypeExpense &&
				spending.AutoCreateTransaction &&
				spending.TargetAmount > 0 {
				if bankAccount == nil {
					bankAccount, err = repo.GetBankAccount(ctx, args.BankAccountId)
					if err != nil {
						log.ErrorContext(
							ctx,
							"failed to retrieve bank account for auto created transaction",
							"err", err,
						)
						return err
					}
				}

				txn := models.Transaction{
					BankAccountId:       spending.BankAccountId,
					Amount:              spending.TargetAmount,
					Date:                dueDate,
					Name:                spending.Name,
					OriginalName:        spending.Name,
					IsPending:           false,
					Source:              models.TransactionSourceManual,
					SpendingId:          &spending.SpendingId,
					CreatedBySpendingId: &spending.SpendingId,
				}

				// AddExpenseToTransaction recalculates the contribution using
				// spending.FundingSchedule, so make sure that relation is set.
				spending.FundingSchedule = fundingSchedule

				if err = repo.AddExpenseToTransaction(ctx, &txn, &spending); err != nil {
					log.ErrorContext(
						ctx,
						"failed to add expense to auto created transaction",
						"err", err,
					)
					return err
				}

				// Always subtract from our available balance. Subtract because credits
				// are represented as negative values in monetr.
				bankAccount.AvailableBalance -= txn.Amount
				bankAccount.CurrentBalance -= txn.Amount

				transactionsToCreate = append(transactionsToCreate, txn)
			}

			spendingToUpdate = append(spendingToUpdate, spending)
		}

		if len(spendingToUpdate) == 0 {
			log.InfoContext(ctx, "no stale spending object were updated")
			return nil
		}

		log.InfoContext(ctx, "updating stale spending objects", "count", len(spendingToUpdate))

		if err = repo.UpdateSpending(
			ctx,
			args.BankAccountId,
			spendingToUpdate,
		); err != nil {
			return errors.Wrap(err, "failed to update stale spending")
		}

		if bankAccount != nil {
			if err = repo.UpdateBankAccount(ctx, bankAccount); err != nil {
				return errors.Wrap(err, "failed to update bank account for auto created transactions")
			}
		}

		for i := range transactionsToCreate {
			if err = repo.CreateTransaction(ctx, args.BankAccountId, &transactionsToCreate[i]); err != nil {
				return errors.Wrap(err, "failed to create auto created transaction")
			}
		}

		return nil
	})
}
