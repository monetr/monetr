//go:build !vault
// +build !vault

package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidLink(ctx context.Context, link *models.PlaidLink) error {
	_, err := r.txn.Model(link).Insert(link)
	return errors.Wrap(err, "failed to create plaid link")
}

func (r *repositoryBase) UpdatePlaidLink(ctx context.Context, link *models.PlaidLink) error {
	span := sentry.StartSpan(ctx, "UpdatePlaidLink")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	span.SetTag("accountId", r.AccountIdStr())
	_, err := r.txn.ModelContext(span.Context(), link).WherePK().Update(link)
	return errors.Wrap(err, "failed to update Plaid link")
}

func (r *repositoryBase) DeletePlaidLink(ctx context.Context, plaidLinkId uint64) error {
	span := sentry.StartSpan(ctx, "DeletePlaidLink")
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &models.Link{}).
		Set(`"plaid_link_id" = NULL`).
		Set(`"link_type" = ?`, models.ManualLinkType).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."plaid_link_id" = ?`, plaidLinkId).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to clean Plaid link prior to removal")
	}

	_, err = r.txn.ModelContext(span.Context(), &models.PlaidLink{}).
		Where(`"plaid_link"."plaid_link_id" = ?`, plaidLinkId).
		ForceDelete()
	return errors.Wrap(err, "failed to delete Plaid link")
}

type PlaidRepository interface {
	GetLinkByItemId(ctx context.Context, itemId string) (*models.Link, error)
	GetLink(ctx context.Context, accountId, linkId uint64) (*models.Link, error)
}

func NewPlaidRepository(db pg.DBI) PlaidRepository {
	return &plaidRepositoryBase{
		txn: db,
	}
}

type plaidRepositoryBase struct {
	txn pg.DBI
}

func (r *plaidRepositoryBase) GetLinkByItemId(ctx context.Context, itemId string) (*models.Link, error) {
	span := sentry.StartSpan(ctx, "GetLinkByItemId")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"itemId": itemId,
	}

	var link models.Link
	err := r.txn.ModelContext(span.Context(), &link).
		Relation("PlaidLink").
		Relation("BankAccounts").
		Where(`"plaid_link"."item_id" = ?`, itemId).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve link by item Id")
	}

	return &link, nil
}

func (r *plaidRepositoryBase) GetLink(ctx context.Context, accountId, linkId uint64) (*models.Link, error) {
	span := sentry.StartSpan(ctx, "GetLink")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": accountId,
		"linkId":    linkId,
	}

	var link models.Link
	err := r.txn.ModelContext(span.Context(), &link).
		Relation("PlaidLink").
		Relation("BankAccounts").
		Where(`"link"."account_id" = ?`, accountId).
		Where(`"link"."link_id" = ?`, linkId).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve link")
	}

	return &link, nil
}
