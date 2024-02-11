package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateTellerLink(ctx context.Context, link *models.TellerLink) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("accountId", r.AccountIdStr())
	span.SetTag("tellerEnrollmentId", link.EnrollmentId)

	link.AccountId = r.AccountId()
	link.UpdatedAt = r.clock.Now().UTC()
	link.CreatedAt = r.clock.Now().UTC()
	link.CreatedByUserId = r.UserId()
	_, err := r.txn.ModelContext(span.Context(), link).Insert(link)
	return errors.Wrap(err, "failed to create Teller link")
}

func (r *repositoryBase) UpdateTellerLink(ctx context.Context, link *models.TellerLink) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("accountId", r.AccountIdStr())
	span.SetTag("tellerEnrollmentId", link.EnrollmentId)

	link.UpdatedAt = r.clock.Now().UTC()

	_, err := r.txn.ModelContext(span.Context(), link).
		WherePK().
		Update(link)
	return errors.Wrap(err, "failed to update Teller link")
}

func (r *repositoryBase) DeleteTellerLink(ctx context.Context, tellerLinkId uint64) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Update the link record to indicate that it is no longer a Teller link but
	// instead a manual one. This way some data is still preserved.
	_, err := r.txn.ModelContext(span.Context(), &models.Link{}).
		Set(`"teller_link_id" = NULL`).
		Set(`"link_type" = ?`, models.ManualLinkType).
		Where(`"link"."account_id" = ?`, r.AccountId()).
		Where(`"link"."teller_link_id" = ?`, tellerLinkId).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to clean link prior to removal")
	}

	// Then delete the Teller link itself.
	_, err = r.txn.ModelContext(span.Context(), &models.TellerLink{}).
		Where(`"teller_link"."teller_link_id" = ?`, tellerLinkId).
		ForceDelete()
	return errors.Wrap(err, "failed to delete Teller link")
}
