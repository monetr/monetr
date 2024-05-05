package repository

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidLink(ctx context.Context, link *PlaidLink) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	link.AccountId = r.AccountId()
	link.CreatedAt = r.clock.Now().UTC()
	link.CreatedBy = r.UserId()
	_, err := r.txn.ModelContext(span.Context(), link).Insert(link)
	return errors.Wrap(err, "failed to create plaid link")
}

func (r *repositoryBase) UpdatePlaidLink(ctx context.Context, link *PlaidLink) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("accountId", r.AccountIdStr())
	span.SetTag("plaidItemId", link.PlaidId)

	_, err := r.txn.ModelContext(span.Context(), link).
		WherePK().
		Update(link)
	return errors.Wrap(err, "failed to update Plaid link")
}

func (r *repositoryBase) DeletePlaidLink(ctx context.Context, plaidLinkId ID[PlaidLink]) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Update the link record to indicate that it is no longer a Plaid link but instead a manual one. This way some data
	// is still preserved.
	_, err := r.txn.ModelContext(span.Context(), &Link{}).
		Set(`"plaid_link_id" = NULL`).
		Set(`"link_type" = ?`, ManualLinkType).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."plaid_link_id" = ?`, plaidLinkId).
		// TODO That won't work anymore
		Where(`"link"."link_type" = ?`, PlaidLinkType).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to clean Plaid link prior to removal")
	}

	// Remove the sync records related to the Plaid link.
	_, err = r.txn.ModelContext(span.Context(), &PlaidSync{}).
		Where(`"plaid_sync"."plaid_link_id" = ?`, plaidLinkId).
		ForceDelete()
	if err != nil {
		return errors.Wrap(err, "failed to clean up sync data for Plaid link removal")
	}

	// Then delete the Plaid link itself.
	_, err = r.txn.ModelContext(span.Context(), &PlaidLink{}).
		Where(`"plaid_link"."plaid_link_id" = ?`, plaidLinkId).
		ForceDelete()
	return errors.Wrap(err, "failed to delete Plaid link")
}

type PlaidRepository interface {
	GetLinkByItemId(ctx context.Context, itemId string) (*Link, error)
	GetLink(ctx context.Context, accountId ID[Account], linkId ID[Link]) (*Link, error)
}

func NewPlaidRepository(db pg.DBI) PlaidRepository {
	return &plaidRepositoryBase{
		txn: db,
	}
}

type plaidRepositoryBase struct {
	txn pg.DBI
}

func (r *plaidRepositoryBase) GetLinkByItemId(ctx context.Context, itemId string) (*Link, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"itemId": itemId,
	}

	var link Link
	err := r.txn.ModelContext(span.Context(), &link).
		Relation("PlaidLink").
		Where(`"plaid_link"."item_id" = ?`, itemId).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve link by item Id")
	}

	return &link, nil
}

func (r *plaidRepositoryBase) GetLink(ctx context.Context, accountId ID[Account], linkId ID[Link]) (*Link, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": accountId,
		"linkId":    linkId,
	}

	var link Link
	err := r.txn.ModelContext(span.Context(), &link).
		Relation("PlaidLink").
		Where(`"link"."account_id" = ?`, accountId).
		Where(`"link"."link_id" = ?`, linkId).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve link")
	}

	return &link, nil
}
