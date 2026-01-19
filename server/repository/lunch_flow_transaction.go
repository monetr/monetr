package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateLunchFlowTransactions(
	ctx context.Context,
	transactions []LunchFlowTransaction,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	now := r.clock.Now()
	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
		transactions[i].CreatedAt = now
	}

	_, err := r.txn.ModelContext(span.Context(), &transactions).Insert(&transactions)
	return errors.Wrap(err, "failed to create Lunch Flow transactions")
}
