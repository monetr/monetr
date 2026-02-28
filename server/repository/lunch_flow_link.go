package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateLunchFlowLink(
	ctx context.Context,
	link *LunchFlowLink,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	now := r.clock.Now().UTC()
	link.AccountId = r.AccountId()
	link.CreatedAt = now
	link.UpdatedAt = now
	link.CreatedBy = r.UserId()
	_, err := r.txn.ModelContext(span.Context(), link).Insert(link)
	return errors.Wrap(err, "failed to create Lunch Flow link")
}

func (r *repositoryBase) UpdateLunchFlowLink(
	ctx context.Context,
	link *LunchFlowLink,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	link.AccountId = r.AccountId()
	link.UpdatedAt = r.clock.Now().UTC()
	_, err := r.txn.ModelContext(span.Context(), link).
		WherePK().
		Update(link)
	return errors.Wrap(err, "failed to update Lunch Flow link")
}

func (r *repositoryBase) RemoveLunchFlowLink(
	ctx context.Context,
	id ID[LunchFlowLink],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Update the link record to indicate that it is no longer a Lunch Flow link
	// but instead a manual one. This way some data is still preserved.
	_, err := r.txn.ModelContext(span.Context(), &Link{}).
		Set(`"link_type" = ?`, ManualLinkType).
		Set(`"lunch_flow_link_id" = NULL`).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."lunch_flow_link_id" = ?`, id).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to clean Lunch Flow link prior to removal")
	}

	// Then delete the Plaid link itself.
	_, err = r.txn.ModelContext(span.Context(), &LunchFlowLink{}).
		Where(`"lunch_flow_link"."account_id" = ?`, r.AccountId()).
		Where(`"lunch_flow_link"."lunch_flow_link_id" = ?`, id).
		Delete()
	return errors.Wrap(err, "failed to delete Lunch Flow link")
}

func (r *repositoryBase) GetLunchFlowLinks(
	ctx context.Context,
) ([]LunchFlowLink, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]LunchFlowLink, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"lunch_flow_link"."account_id" = ?`, r.AccountId()).
		Where(`"lunch_flow_link"."deleted_at" IS NULL`).
		Order(`lunch_flow_link_id DESC`).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Lunch Flow links")
	}

	return result, nil
}

func (r *repositoryBase) GetLunchFlowLink(
	ctx context.Context,
	id ID[LunchFlowLink],
) (*LunchFlowLink, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var link LunchFlowLink
	err := r.txn.ModelContext(span.Context(), &link).
		Where(`"lunch_flow_link"."account_id" = ?`, r.AccountId()).
		Where(`"lunch_flow_link"."lunch_flow_link_id" = ?`, id).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve Lunch Flow link")
	}

	return &link, nil
}
