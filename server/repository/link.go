package repository

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetLink(ctx context.Context, linkId uint64) (*models.Link, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"linkId": linkId,
	}

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

func (r *repositoryBase) GetLinks(ctx context.Context) ([]models.Link, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result []models.Link
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"link"."account_id" = ?`, r.accountId).
		Select(&result)
	if err != nil {
		return nil, crumbs.WrapError(span.Context(), err, "failed to retrieve links")
	}

	return result, nil
}

func (r *repositoryBase) GetNumberOfPlaidLinks(ctx context.Context) (int, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	count, err := r.txn.ModelContext(span.Context(), &models.Link{}).
		Where(`"link"."account_id" = ?`, r.accountId).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Count()
	if err != nil {
		return count, crumbs.WrapError(span.Context(), err, "failed to retrieve links")
	}

	return count, nil
}

func (r *repositoryBase) GetLinkIsManual(ctx context.Context, linkId uint64) (bool, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"linkId": linkId,
	}

	ok, err := r.txn.ModelContext(span.Context(), &models.Link{}).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."link_id" = ?`, linkId).
		Where(`"link"."link_type" = ?`, models.ManualLinkType).
		Exists()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, crumbs.WrapError(span.Context(), err, "failed to get link is manual")
	}

	span.Status = sentry.SpanStatusOK

	return ok, nil
}

func (r *repositoryBase) GetLinkIsManualByBankAccountId(ctx context.Context, bankAccountId uint64) (bool, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"bankAccountId": bankAccountId,
	}

	ok, err := r.txn.ModelContext(span.Context(), &models.Link{}).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."link_id" = "link"."link_id" AND "bank_account"."account_id" = "link"."account_id"`).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."bank_account_id" = ?`, bankAccountId).
		Where(`"link"."link_type" = ?`, models.ManualLinkType).
		Exists()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, crumbs.WrapError(span.Context(), err, "failed to get link by bank account Id")
	}

	span.Status = sentry.SpanStatusOK

	return ok, nil
}

func (r *repositoryBase) CreateLink(ctx context.Context, link *models.Link) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	userId := r.UserId()
	now := time.Now().UTC()
	link.AccountId = r.AccountId()
	link.CreatedByUserId = userId
	link.CreatedAt = now
	link.UpdatedAt = now

	_, err := r.txn.ModelContext(span.Context(), link).Insert(link)
	return errors.Wrap(err, "failed to insert link")
}

func (r *repositoryBase) UpdateLink(ctx context.Context, link *models.Link) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"link": *link,
	}

	link.AccountId = r.AccountId()
	link.UpdatedAt = time.Now().UTC()

	_, err := r.txn.ModelContext(span.Context(), link).
		WherePK().
		Returning(`*`).
		Update(link)
	return errors.Wrap(err, "failed to update link")
}

func (r *repositoryBase) UpdateLinkManualSyncTimestampMaybe(ctx context.Context, linkId uint64) (ok bool, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"linkId": linkId,
	}

	var link models.Link
	now := time.Now()
	result, err := r.txn.ModelContext(span.Context(), &link).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."link_id" = ?`, linkId).
		// Only let them manually sync once every 30 minutes.
		Where(`("link"."last_manual_sync" < ? OR "link"."last_manual_sync" IS NULL)`, now.Add(-30*time.Minute)).
		Set(`"last_manual_sync" = ?`, now).
		Update()
	if err != nil {
		return false, errors.Wrap(err, "failed to update last manual sync timestamp")
	}

	// If we actually updated a row then that means we can proceed with the syncing.
	if result.RowsAffected() == 1 {
		return true, nil
	}

	// If we didn't update the row then that means we cannot do a manual resync right now.
	return false, nil
}
