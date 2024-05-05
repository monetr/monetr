package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidBankAccount(
	ctx context.Context,
	bankAccount *models.PlaidBankAccount,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	bankAccount.AccountId = r.AccountId()
	bankAccount.CreatedAt = r.clock.Now().UTC()
	bankAccount.CreatedBy = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), bankAccount).Insert(bankAccount)

	return errors.Wrap(err, "failed to create plaid bank account")
}

func (r *repositoryBase) UpdatePlaidBankAccount(
	ctx context.Context,
	bankAccount *models.PlaidBankAccount,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	bankAccount.AccountId = r.AccountId()

	_, err := r.txn.ModelContext(span.Context(), bankAccount).
		WherePK().
		Update(bankAccount)

	return errors.Wrap(err, "failed to update plaid bank account")
}
