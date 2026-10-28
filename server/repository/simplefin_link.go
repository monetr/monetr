package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateSimpleFinLink(
	ctx context.Context,
	link *SimpleFinLink,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())

	now := r.clock.Now().UTC()
	link.AccountId = r.AccountId()
	link.CreatedAt = now
	link.CreatedBy = r.UserId()
	link.UpdatedAt = now

	if _, err := r.txn.NewInsert().
		Model(link).
		Returning("*").
		Exec(span.Context()); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create SimpleFIN link")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) UpdateSimpleFinLink(
	ctx context.Context,
	link *SimpleFinLink,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("simpleFinLinkId", link.SimpleFinLinkId)

	link.AccountId = r.AccountId()
	link.UpdatedAt = r.clock.Now().UTC()

	result, err := r.txn.NewUpdate().Model(link).
		WherePK().
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update SimpleFIN link")
	} else if affected, err := result.RowsAffected(); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update SimpleFIN link")
	} else if affected != 1 {
		span.Status = sentry.SpanStatusNotFound
		return errors.New("no rows updated")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) GetSimpleFinLinks(
	ctx context.Context,
) ([]SimpleFinLink, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())

	result := make([]SimpleFinLink, 0)
	err := r.txn.NewSelect().Model(&result).
		Where(`"simplefin_link"."account_id" = ?`, r.AccountId()).
		Where(`"simplefin_link"."deleted_at" IS NULL`).
		Order(`simplefin_link_id DESC`).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve SimpleFIN links")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetSimpleFinLink(
	ctx context.Context,
	id ID[SimpleFinLink],
) (*SimpleFinLink, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("simpleFinLinkId", id)

	var result SimpleFinLink
	err := r.txn.NewSelect().Model(&result).
		Where(`"simplefin_link"."account_id" = ?`, r.AccountId()).
		Where(`"simplefin_link"."simplefin_link_id" = ?`, id).
		Limit(1).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve SimpleFIN link")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}
