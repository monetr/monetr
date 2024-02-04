package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateTellerBankAccount(ctx context.Context, bankAccount *models.TellerBankAccount) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("accountId", r.AccountIdStr())

	bankAccount.AccountId = r.AccountId()
	bankAccount.UpdatedAt = r.clock.Now().UTC()
	bankAccount.CreatedAt = r.clock.Now().UTC()
	_, err := r.txn.ModelContext(span.Context(), bankAccount).Insert(bankAccount)
	return errors.Wrap(err, "failed to create Teller bank account")
}

func (r *repositoryBase) UpdateTellerBankAccount(ctx context.Context, bankAccount *models.TellerBankAccount) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("accountId", r.AccountIdStr())

	bankAccount.UpdatedAt = r.clock.Now().UTC()

	_, err := r.txn.ModelContext(span.Context(), bankAccount).
		WherePK().
		Update(bankAccount)
	return errors.Wrap(err, "failed to update Teller bank account")
}
