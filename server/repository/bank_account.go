package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetBankAccounts(ctx context.Context) ([]BankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
	}

	result := make([]BankAccount, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("PlaidBankAccount").
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Select(&result)
	return result, errors.Wrap(err, "failed to retrieve bank accounts")
}

func (r *repositoryBase) CreateBankAccounts(ctx context.Context, bankAccounts ...*BankAccount) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
	}

	now := r.clock.Now().UTC()
	for i := range bankAccounts {
		bankAccounts[i].AccountId = r.AccountId()
		bankAccounts[i].CreatedAt = now
		if bankAccounts[i].Status == "" {
			bankAccounts[i].Status = ActiveBankAccountStatus
		}
	}
	if _, err := r.txn.ModelContext(span.Context(), &bankAccounts).Insert(&bankAccounts); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to insert bank accounts")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) GetBankAccountsByLinkId(ctx context.Context, linkId ID[Link]) ([]BankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
		"linkId":    linkId,
	}

	var result []BankAccount
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("PlaidBankAccount").
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

func (r *repositoryBase) GetBankAccountsWithPlaidByLinkId(
	ctx context.Context,
	linkId ID[Link],
) ([]BankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
		"linkId":    linkId,
	}

	var result []BankAccount
	err := r.txn.ModelContext(span.Context(), &result).
		Relation(`PlaidBankAccount`).
		Where(`"bank_account"."plaid_bank_account_id" IS NOT NULL`).
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

func (r *repositoryBase) GetPlaidBankAccountsByLinkId(ctx context.Context, linkId ID[Link]) ([]PlaidBankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
		"linkId":    linkId,
	}

	var result []PlaidBankAccount
	err := r.txn.ModelContext(span.Context(), &result).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"link"."account_id" = "plaid_bank_account"."account_id" AND "link"."plaid_link_id" = "plaid_bank_account"."plaid_link_id"`).
		Where(`"plaid_bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"link"."link_id" = ? `, linkId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve bank accounts by Id")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetBankAccount(ctx context.Context, bankAccountId ID[BankAccount]) (*BankAccount, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

	var result BankAccount
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("PlaidBankAccount").
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

func (r *repositoryBase) UpdateBankAccounts(ctx context.Context, accounts ...BankAccount) error {
	if len(accounts) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Make sure each of the accounts has the correct accountId.
	bankAccountIds := make([]ID[BankAccount], len(accounts))
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
