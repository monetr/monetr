//go:build !vault
// +build !vault

package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func (r *repositoryBase) CreatePlaidLink(ctx context.Context, link *models.PlaidLink) error {
	span := sentry.StartSpan(ctx, "CreatePlaidLink")
	defer span.Finish()

	_, err := r.db.NewInsert().Model(link).Exec(span.Context(), link)
	return errors.Wrap(err, "failed to create plaid link")
}

func (r *repositoryBase) UpdatePlaidLink(ctx context.Context, link *models.PlaidLink) error {
	span := sentry.StartSpan(ctx, "UpdatePlaidLink")
	defer span.Finish()

	_, err := r.db.NewUpdate().Model(link).WherePK().Exec(span.Context(), link)
	return errors.Wrap(err, "failed to update Plaid link")
}

type PlaidRepository interface {
	GetLinkByItemId(ctx context.Context, itemId string) (*models.Link, error)
	GetLink(ctx context.Context, accountId, linkId uint64) (*models.Link, error)
}

func NewPlaidRepository(db bun.IDB) PlaidRepository {
	return &plaidRepositoryBase{
		db: db,
	}
}

type plaidRepositoryBase struct {
	db bun.IDB
}

func (r *plaidRepositoryBase) GetLinkByItemId(ctx context.Context, itemId string) (*models.Link, error) {
	span := sentry.StartSpan(ctx, "GetLinkByItemId")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"itemId": itemId,
	}

	var link models.Link
	err := r.db.NewSelect().
		Model(&link).
		Relation("PlaidLink").
		Relation("BankAccounts").
		Where(`plaid_link.item_id = ?`, itemId).
		Limit(1).
		Scan(span.Context(), &link)
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
	err := r.db.NewSelect().
		Model(&link).
		Relation("PlaidLink").
		Relation("BankAccounts").
		Where(`link.account_id = ?`, accountId).
		Where(`link.link_id = ?`, linkId).
		Limit(1).
		Scan(span.Context(), &link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve link")
	}

	return &link, nil
}
