package recurring_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
)

type CalculateRecurringTransactionsArguments struct {
	AccountId            models.ID[models.Account]            `json:"accountId"`
	BankAccountId        models.ID[models.BankAccount]        `json:"bankAccountId"`
	TransactionClusterId models.ID[models.TransactionCluster] `json:"transactionClusterId"`
}

func CalculateRecurringTransactions(ctx queue.Context, args CalculateRecurringTransactionsArguments) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		crumbs.IncludeUserInScope(ctx, args.AccountId)
		// log := ctx.Log().With(
		// 	"accountId", args.AccountId,
		// 	"bankAccountId", args.BankAccountId,
		// )

		// repo := repository.NewRepositoryFromSession(
		// 	ctx.Clock(),
		// 	"user_system",
		// 	args.AccountId,
		// 	ctx.DB(),
		// 	log,
		// )

		return nil
	})
}
