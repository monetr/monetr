package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateSimpleFinBankAccount(
	ctx context.Context,
	bankAccount *SimpleFinBankAccount,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("simpleFinLinkId", bankAccount.SimpleFinLinkId)

	now := r.clock.Now().UTC()
	bankAccount.AccountId = r.AccountId()
	bankAccount.CreatedAt = now
	bankAccount.CreatedBy = r.UserId()
	bankAccount.UpdatedAt = now

	if _, err := r.txn.NewInsert().
		Model(bankAccount).
		Returning("*").
		Exec(span.Context()); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create SimpleFIN bank account")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) UpdateSimpleFinBankAccount(
	ctx context.Context,
	bankAccount *SimpleFinBankAccount,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("simpleFinBankAccountId", bankAccount.SimpleFinBankAccountId)

	bankAccount.AccountId = r.AccountId()
	bankAccount.UpdatedAt = r.clock.Now().UTC()

	result, err := r.txn.NewUpdate().Model(bankAccount).
		WherePK().
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update SimpleFIN bank account")
	} else if affected, err := result.RowsAffected(); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update SimpleFIN bank account")
	} else if affected != 1 {
		span.Status = sentry.SpanStatusNotFound
		return errors.New("no rows updated")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) GetSimpleFinBankAccountForSimpleFinLink(
	ctx context.Context,
	linkId ID[SimpleFinLink],
	id ID[SimpleFinBankAccount],
) (*SimpleFinBankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("simpleFinLinkId", linkId)
	span.SetData("simpleFinBankAccountId", id)

	var result SimpleFinBankAccount
	err := r.txn.NewSelect().Model(&result).
		Where(`"simplefin_bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"simplefin_bank_account"."simplefin_link_id" = ?`, linkId).
		Where(`"simplefin_bank_account"."simplefin_bank_account_id" = ?`, id).
		Where(`"simplefin_bank_account"."deleted_at" IS NULL`).
		Limit(1).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve SimpleFIN bank account")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (r *repositoryBase) GetSimpleFinBankAccountsBySimpleFinLink(
	ctx context.Context,
	linkId ID[SimpleFinLink],
) ([]SimpleFinBankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("simpleFinLinkId", linkId)

	result := make([]SimpleFinBankAccount, 0)
	err := r.txn.NewSelect().Model(&result).
		Where(`"simplefin_bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"simplefin_bank_account"."simplefin_link_id" = ?`, linkId).
		Where(`"simplefin_bank_account"."deleted_at" IS NULL`).
		Order(`simplefin_bank_account_id DESC`).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve SimpleFIN bank accounts")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

// GetAllSimpleFinBankAccountsBySimpleFinLink is the same as
// [GetSimpleFinBankAccountsBySimpleFinLink] except it does not exclude deleted
// accounts. This is used when matching what SimpleFIN has against what we
// have, a deleted account still holds its spot in the unique constraint so we
// can't create it again.
func (r *repositoryBase) GetAllSimpleFinBankAccountsBySimpleFinLink(
	ctx context.Context,
	linkId ID[SimpleFinLink],
) ([]SimpleFinBankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("simpleFinLinkId", linkId)

	result := make([]SimpleFinBankAccount, 0)
	err := r.txn.NewSelect().Model(&result).
		Where(`"simplefin_bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"simplefin_bank_account"."simplefin_link_id" = ?`, linkId).
		Order(`simplefin_bank_account_id DESC`).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve SimpleFIN bank accounts")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}
