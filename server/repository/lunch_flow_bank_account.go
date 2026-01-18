package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateLunchFlowBankAccount(
	ctx context.Context,
	bankAccount *LunchFlowBankAccount,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	now := r.clock.Now()
	bankAccount.AccountId = r.AccountId()
	bankAccount.CreatedAt = now
	bankAccount.UpdatedAt = now
	bankAccount.CreatedBy = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), bankAccount).Insert(bankAccount)
	return errors.Wrap(err, "failed to create Lunch Flow Bank Account")
}

func (r *repositoryBase) GetLunchFlowBankAccountsByLunchFlowLink(
	ctx context.Context, id ID[LunchFlowLink],
) ([]LunchFlowBankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]LunchFlowBankAccount, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"lunch_flow_bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"lunch_flow_bank_account"."lunch_flow_link_id" = ?`, id).
		Where(`"lunch_flow_bank_account"."deleted_at" IS NULL`).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Lunch Flow bank account")
	}

	return result, nil
}
