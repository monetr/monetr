package recurring_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

const (
	// matchRecurringTransactionsToSpendingLookback is how many of the most recent
	// transactions of a recurring transaction are considered when matching it to
	// an expense.
	matchRecurringTransactionsToSpendingLookback = 3
	// matchRecurringTransactionsToSpendingMinimumVotes is how many of those
	// transactions must have been spent from the same expense for it to be
	// matched.
	matchRecurringTransactionsToSpendingMinimumVotes = 2
)

type MatchRecurringTransactionsToSpendingArguments struct {
	AccountId     models.ID[models.Account]     `json:"accountId"`
	BankAccountId models.ID[models.BankAccount] `json:"bankAccountId"`
}

// MatchRecurringTransactionsToSpending links recurring transactions that are
// not linked to anything yet to an existing expense. A recurring transaction is
// matched when enough of its most recent transactions were spent from the same
// expense by the user. Recurring transactions that are already linked are never
// changed, and links made here are marked as auto matched.
func MatchRecurringTransactionsToSpending(
	ctx queue.Context,
	args MatchRecurringTransactionsToSpendingArguments,
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

		recurrings, err := repo.GetTransactionRecurringByBankAccount(
			ctx,
			args.BankAccountId,
		)
		if err != nil {
			return errors.Wrap(err, "failed to read recurring transactions")
		}

		spending, err := repo.GetSpending(ctx, args.BankAccountId)
		if err != nil {
			return errors.Wrap(err, "failed to read spending")
		}

		expenses := make(map[models.ID[models.Spending]]bool, len(spending))
		for _, item := range spending {
			if item.SpendingType == models.SpendingTypeExpense {
				expenses[item.SpendingId] = true
			}
		}

		// A spending object can only be linked to a single recurring transaction,
		// so anything already linked is off the table.
		linked := make(map[models.ID[models.Spending]]bool, len(recurrings))
		for _, recurring := range recurrings {
			if recurring.SpendingId != nil {
				linked[*recurring.SpendingId] = true
			}
		}

		matched := 0
		for i := range recurrings {
			recurring := &recurrings[i]
			if recurring.SpendingId != nil ||
				recurring.FundingScheduleId != nil ||
				recurring.Ended ||
				recurring.Direction != models.DebitDirection {
				continue
			}

			transactions, err := repo.GetTransactionsForRecurring(
				ctx,
				args.BankAccountId,
				recurring.TransactionRecurringId,
				matchRecurringTransactionsToSpendingLookback,
				0,
			)
			if err != nil {
				return errors.Wrap(err, "failed to read transactions for recurring transaction")
			}

			spendingId, ok := matchSpendingForRecurring(transactions)
			if !ok || !expenses[spendingId] || linked[spendingId] {
				continue
			}

			recurring.SpendingId = &spendingId
			recurring.AutoMatched = true
			if err := repo.UpdateTransactionRecurring(
				ctx,
				args.BankAccountId,
				recurring,
			); err != nil {
				return errors.Wrap(err, "failed to link recurring transaction to expense")
			}

			log.DebugContext(ctx, "matched recurring transaction to expense",
				"transactionRecurringId", recurring.TransactionRecurringId,
				"spendingId", spendingId,
			)
			linked[spendingId] = true
			matched++
		}

		log.InfoContext(ctx, "finished matching recurring transactions to spending",
			"recurring", len(recurrings),
			"matched", matched,
		)

		return nil
	})
}

// matchSpendingForRecurring returns the spending that enough of the provided
// transactions were spent from. If no spending has enough transactions then
// false is returned.
func matchSpendingForRecurring(
	transactions []models.Transaction,
) (models.ID[models.Spending], bool) {
	votes := make(map[models.ID[models.Spending]]int, len(transactions))
	for _, transaction := range transactions {
		if transaction.SpendingId == nil {
			continue
		}

		spendingId := *transaction.SpendingId
		votes[spendingId]++
		if votes[spendingId] >= matchRecurringTransactionsToSpendingMinimumVotes {
			return spendingId, true
		}
	}

	return "", false
}
