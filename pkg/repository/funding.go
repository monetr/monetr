package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) UpdateSpendingFunding(
	ctx context.Context,
	bankAccountId uint64,
	updates []models.SpendingFunding,
) error {
	if len(updates) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	spendingFundingIds := make([]uint64, len(updates))
	for i := range updates {
		updates[i].AccountId = r.AccountId()
		updates[i].BankAccountId = bankAccountId
		spendingFundingIds[i] = updates[i].SpendingFundingId
	}

	span.Data = map[string]interface{}{
		"accountId":          r.AccountId(),
		"bankAccountId":      bankAccountId,
		"spendingFundingIds": spendingFundingIds,
	}

	_, err := r.txn.ModelContext(span.Context(), &updates).
		Update(&updates)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update expenses")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) GetSpendingFunding(ctx context.Context, bankAccountId uint64, spendingId uint64) ([]models.SpendingFunding, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]models.SpendingFunding, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("FundingSchedule").
		Where(`"spending_funding"."account_id" = ?`, r.AccountId()).
		Where(`"spending_funding"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending_funding"."spending_id" = ?`, spendingId).
		Select(&result)
	return result, errors.Wrap(err, "failed to retrieve spending funding")
}

func (r *repositoryBase) CreateSpendingFunding(ctx context.Context, funding []models.SpendingFunding) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	for i, item := range funding {
		myownsanity.Assert(item.BankAccountId > 0, "SpendingFunding bank account ID must be set by the caller")
		funding[i].AccountId = r.AccountId()
	}

	_, err := r.txn.ModelContext(span.Context(), &funding).Insert(&funding)
	return errors.Wrap(err, "failed to create spending funding")
}
