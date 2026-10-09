package recurring_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
)

type AutoAssignRecurringTransactionsArguments struct {
	AccountId      models.ID[models.Account]       `json:"accountId"`
	BankAccountId  models.ID[models.BankAccount]   `json:"bankAccountId"`
	TransactionIds []models.ID[models.Transaction] `json:"transactionIds"`
}

// AutoAssignRecurringTransactions will spend the provided transactions from the
// spending their recurring transaction is linked to, as long as the user turned
// on auto assign for it. The transactions should be the ones that were just
// added to a recurring transaction.
func AutoAssignRecurringTransactions(
	ctx queue.Context,
	args AutoAssignRecurringTransactionsArguments,
) error {
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
			return errors.Wrap(err, "failed to read account")
		}
		timezone, err := account.GetTimezone()
		if err != nil {
			return errors.Wrap(err, "failed to parse account timezone")
		}

		recurrings, err := repo.GetTransactionRecurringByBankAccount(
			ctx,
			args.BankAccountId,
		)
		if err != nil {
			return errors.Wrap(err, "failed to read recurring transactions")
		}
		recurringById := make(
			map[models.ID[models.TransactionRecurring]]models.TransactionRecurring,
			len(recurrings),
		)
		for _, item := range recurrings {
			recurringById[item.TransactionRecurringId] = item
		}

		spending, err := repo.GetSpending(ctx, args.BankAccountId)
		if err != nil {
			return errors.Wrap(err, "failed to read spending")
		}
		spendingById := make(
			map[models.ID[models.Spending]]models.Spending,
			len(spending),
		)
		for _, item := range spending {
			spendingById[item.SpendingId] = item
		}

		// TODO This does a few queries per transaction, eventually this should be
		// one bulk update for spending and one for transactions.
		var assigned int
		for _, transactionId := range args.TransactionIds {
			transaction, err := repo.GetTransaction(
				ctx,
				args.BankAccountId,
				transactionId,
			)
			if err != nil {
				return errors.Wrap(err, "failed to read transaction")
			}

			// If the transaction is already spent from something then leave it alone,
			// whatever it is spent from now was picked by the user or already
			// assigned.
			if transaction.SpendingId != nil ||
				transaction.DeletedAt != nil ||
				transaction.TransactionRecurringId == nil {
				continue
			}

			recurring, ok := recurringById[*transaction.TransactionRecurringId]
			if !ok || !recurring.AutoAssign || recurring.SpendingId == nil {
				continue
			}

			item, ok := spendingById[*recurring.SpendingId]
			if !ok {
				continue
			}

			// Don't spend transactions that happened before the spending even
			// existed, those would have been paid for some other way.
			if transaction.Date.Before(util.Midnight(item.CreatedAt, timezone)) {
				log.DebugContext(ctx,
					"skipping auto assing spending to transaction because transaction date is after spending was created",
					"transactionId", transactionId,
					"transactionRecurringId", recurring.TransactionRecurringId,
					"spendingId", *recurring.SpendingId,
					"transaction_date", transaction.Date,
					"spending_createdAt", item.CreatedAt,
					"spending_createdAt_adjusted", util.Midnight(item.CreatedAt, timezone),
				)
				continue
			}

			updated := *transaction
			updated.SpendingId = recurring.SpendingId
			if _, err := repo.ProcessTransactionSpentFrom(
				ctx,
				args.BankAccountId,
				&updated,
				transaction,
			); err != nil {
				return errors.Wrap(err, "failed to spend transaction from spending")
			}

			if err := repo.UpdateTransaction(
				ctx,
				args.BankAccountId,
				&updated,
			); err != nil {
				return errors.Wrap(err, "failed to update transaction")
			}

			log.DebugContext(ctx,
				"auto assigned transaction to spending",
				"transactionId", transactionId,
				"transactionRecurringId", recurring.TransactionRecurringId,
				"spendingId", *recurring.SpendingId,
			)
			assigned++
		}

		log.InfoContext(ctx,
			"finished auto assigning recurring transactions",
			"transactions", len(args.TransactionIds),
			"assigned", assigned,
		)

		return nil
	})
}
