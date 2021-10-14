package repository

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetLink(ctx context.Context, linkId uint64) (*models.Link, error) {
	span := sentry.StartSpan(ctx, "Get Link")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"linkId": linkId,
	}

	var link models.Link
	err := r.db.NewSelect().
		Model(&link).
		Relation("PlaidLink").
		Relation("BankAccounts").
		Where(`link.link_id = ? AND link.account_id = ?`, linkId, r.AccountId()).
		Limit(1).
		Scan(span.Context(), &link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get link")
	}

	return &link, nil
}

func (r *repositoryBase) GetLinks(ctx context.Context) ([]models.Link, error) {
	span := sentry.StartSpan(ctx, "GetLinks")
	defer span.Finish()

	var result []models.Link
	err := r.db.NewSelect().
		Model(&result).
		Where(`link.account_id = ?`, r.AccountId()).
		Scan(span.Context(), &result)
	if err != nil {
		return nil, crumbs.WrapError(span.Context(), err, "failed to retrieve links")
	}

	return result, nil
}

func (r *repositoryBase) GetNumberOfPlaidLinks(ctx context.Context) (int, error) {
	span := sentry.StartSpan(ctx, "GetLinks")
	defer span.Finish()

	count, err := r.db.NewSelect().
		Model(&models.Link{}).
		Where(`link.account_id = ?`, r.AccountId()).
		Where(`link.link_type = ?`, models.PlaidLinkType).
		Count(span.Context())
	if err != nil {
		return count, crumbs.WrapError(span.Context(), err, "failed to retrieve links")
	}

	return count, nil
}

func (r *repositoryBase) GetLinkIsManual(ctx context.Context, linkId uint64) (bool, error) {
	span := sentry.StartSpan(ctx, "GetLinkIsManual")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"linkId": linkId,
	}

	ok, err := r.db.NewSelect().
		Model(&models.Link{}).
		Where(`link.account_id = ?`, r.AccountId()).
		Where(`link.link_id = ?`, linkId).
		Where(`link.link_type = ?`, models.ManualLinkType).
		Exists(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, crumbs.WrapError(span.Context(), err, "failed to get link is manual")
	}

	span.Status = sentry.SpanStatusOK

	return ok, nil
}

func (r *repositoryBase) GetLinkIsManualByBankAccountId(ctx context.Context, bankAccountId uint64) (bool, error) {
	span := sentry.StartSpan(ctx, "GetLinkIsManualByBankAccountId")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"bankAccountId": bankAccountId,
	}

	ok, err := r.db.NewSelect().
		Model(&models.Link{}).
		Join(`INNER JOIN bank_accounts AS bank_account`).
		JoinOn(`bank_account.link_id = link.link_id AND bank_account.account_id = link.account_id`).
		Where(`link.account_id = ?`, r.AccountId()).
		Where(`bank_account.bank_account_id = ?`, bankAccountId).
		Where(`link.link_type = ?`, models.ManualLinkType).
		Exists(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, crumbs.WrapError(span.Context(), err, "failed to get link by bank account Id")
	}

	span.Status = sentry.SpanStatusOK

	return ok, nil
}

func (r *repositoryBase) CreateLink(ctx context.Context, link *models.Link) error {
	span := sentry.StartSpan(ctx, "CreateLink")
	defer span.Finish()

	userId := r.UserId()
	now := time.Now().UTC()
	link.AccountId = r.AccountId()
	link.CreatedByUserId = userId
	link.CreatedAt = now
	link.UpdatedAt = now

	_, err := r.db.NewInsert().Model(link).Exec(span.Context(), link)
	return errors.Wrap(err, "failed to insert link")
}

func (r *repositoryBase) UpdateLink(ctx context.Context, link *models.Link) error {
	span := sentry.StartSpan(ctx, "UpdateLink")
	defer span.Finish()

	link.AccountId = r.AccountId()
	link.UpdatedAt = time.Now().UTC()

	// TODO might need to add stuff for returning.
	_, err := r.db.NewUpdate().Model(link).WherePK().OmitZero().Exec(span.Context(), &link)
	return errors.Wrap(err, "failed to update link")
}
