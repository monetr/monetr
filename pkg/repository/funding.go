package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) UpdateSpendingFunding(
	ctx context.Context,
	bankAccountId uint64,
	updates []models.SpendingFunding,
) error {
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
