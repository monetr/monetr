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

func (r *repositoryBase) DeletePlaidLink(
	ctx context.Context,
	plaidLinkId ID[PlaidLink],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	now := r.clock.Now()

	// Update the link record to indicate that it is no longer a Plaid link but
	// instead a manual one. This way some data is still preserved.
	_, err := r.txn.ModelContext(span.Context(), &Link{}).
		Set(`"link_type" = ?`, ManualLinkType).
		Set(`"deleted_at" = ?`, now).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."plaid_link_id" = ?`, plaidLinkId).
		Where(`"link"."link_type" = ?`, PlaidLinkType).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to clean Plaid link prior to removal")
	}

	// Then delete the Plaid link itself.
	_, err = r.txn.ModelContext(span.Context(), &PlaidLink{}).
		Set(`"status" = ?`, PlaidLinkStatusDeactivated).
		Set(`"deleted_at" = ?`, now).
		Where(`"plaid_link"."account_id" = ?`, r.AccountId()).
		Where(`"plaid_link"."plaid_link_id" = ?`, plaidLinkId).
		Update()
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
	span.Data = map[string]any{
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

	span.Data = map[string]any{
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
