package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetBankAccounts(ctx context.Context) ([]models.BankAccount, error) {
	span := sentry.StartSpan(ctx, "GetBankAccounts")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
	}

	var result []models.BankAccount
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Select(&result)
	return result, errors.Wrap(err, "failed to retrieve bank accounts")
}

func (r *repositoryBase) CreateBankAccounts(ctx context.Context, bankAccounts ...models.BankAccount) error {
	span := sentry.StartSpan(ctx, "CreateBankAccounts")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
	}

	for i := range bankAccounts {
		bankAccounts[i].BankAccountId = 0
		bankAccounts[i].AccountId = r.AccountId()
	}
	if _, err := r.txn.ModelContext(span.Context(), &bankAccounts).Insert(&bankAccounts); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to insert bank accounts")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) GetBankAccountsByLinkId(ctx context.Context, linkId uint64) ([]models.BankAccount, error) {
	span := sentry.StartSpan(ctx, "GetBankAccountsByLinkId")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
		"linkId":    linkId,
	}

	var result []models.BankAccount
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ? `, linkId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve bank accounts by Id")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetBankAccount(ctx context.Context, bankAccountId uint64) (*models.BankAccount, error) {
	span := sentry.StartSpan(ctx, "GetBankAccount")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

	var result models.BankAccount
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."bank_account_id" = ? `, bankAccountId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve bank account")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (r *repositoryBase) UpdateBankAccounts(ctx context.Context, accounts []models.BankAccount) error {
	if len(accounts) == 0 {
		return nil
	}

	span := sentry.StartSpan(ctx, "UpdateBankAccounts")
	defer span.Finish()

	// Make sure each of the accounts has the correct accountId.
	bankAccountIds := make([]uint64, len(accounts))
	for i := range accounts {
		accounts[i].AccountId = r.AccountId()
		bankAccountIds[i] = accounts[i].BankAccountId
	}

	span.Data = map[string]interface{}{
		"accountId":      r.AccountId(),
		"bankAccountIds": bankAccountIds,
	}

	_, err := r.txn.ModelContext(span.Context(), &accounts).
		WherePK().
		UpdateNotZero(&accounts)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update bank accounts")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}
