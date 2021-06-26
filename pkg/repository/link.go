package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
	"time"
)

func (r *repositoryBase) GetLink(ctx context.Context, linkId uint64) (*models.Link, error) {
	span := sentry.StartSpan(ctx, "Get Link")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	span.SetTag("accountId", r.AccountIdStr())

	span.Data["linkId"] = linkId

	var link models.Link
	err := r.txn.ModelContext(span.Context(), &link).
		Relation("PlaidLink").
		Relation("BankAccounts").
		Where(`"link"."link_id" = ? AND "link"."account_id" = ?`, linkId, r.AccountId()).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get link")
	}

	return &link, nil
}

func (r *repositoryBase) GetLinks() ([]models.Link, error) {
	var result []models.Link
	err := r.txn.Model(&result).
		Where(`"link"."account_id" = ?`, r.accountId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve links")
	}

	return result, nil
}

func (r *repositoryBase) GetLinkIsManual(linkId uint64) (bool, error) {
	ok, err := r.txn.Model(&models.Link{}).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."link_id" = ?`, linkId).
		Where(`"link"."link_type" = ?`, models.ManualLinkType).
		Exists()
	if err != nil {
		return false, errors.Wrap(err, "failed to get link by bank account Id")
	}

	return ok, nil
}

func (r *repositoryBase) GetLinkIsManualByBankAccountId(bankAccountId uint64) (bool, error) {
	ok, err := r.txn.Model(&models.Link{}).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."link_id" = "link"."link_id" AND "bank_account"."account_id" = "link"."account_id"`).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."bank_account_id" = ?`, bankAccountId).
		Where(`"link"."link_type" = ?`, models.ManualLinkType).
		Exists()
	if err != nil {
		return false, errors.Wrap(err, "failed to get link by bank account Id")
	}

	return ok, nil
}

func (r *repositoryBase) CreateLink(link *models.Link) error {
	userId := r.UserId()
	now := time.Now().UTC()
	link.AccountId = r.AccountId()
	link.CreatedByUserId = userId
	link.UpdatedByUserId = &userId
	link.CreatedAt = now
	link.UpdatedAt = now

	_, err := r.txn.Model(link).Insert(link)
	return errors.Wrap(err, "failed to insert link")
}

func (r *repositoryBase) UpdateLink(link *models.Link) error {
	userId := r.UserId()
	link.AccountId = r.AccountId()
	link.UpdatedByUserId = &userId
	link.UpdatedAt = time.Now().UTC()

	_, err := r.txn.Model(link).WherePK().Returning(`*`).UpdateNotZero(link)
	return errors.Wrap(err, "failed to update link")
}
