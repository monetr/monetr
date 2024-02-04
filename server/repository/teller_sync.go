package repository

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateTellerSync(ctx context.Context, sync *models.TellerSync) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("accountId", r.AccountIdStr())

	sync.AccountId = r.AccountId()
	_, err := r.txn.ModelContext(span.Context(), sync).Insert(sync)
	return errors.Wrap(err, "failed to create Teller sync")
}

// GetLatestTellerSync will return the latest teller sync by it's timestamp.
// This is used to make sure we only work with a small set of transactions in
// the reconciliation process that is part of the sync by only syncing
// transactions after the immutable timestamp.
func (r *repositoryBase) GetLatestTellerSync(ctx context.Context, tellerBankAccountId uint64) (*models.TellerSync, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.SetTag("accountId", r.AccountIdStr())

	var sync models.TellerSync
	err := r.txn.ModelContext(span.Context(), &sync).
		Where(`"teller_sync"."account_id" = ?`, r.AccountId()).
		Where(`"teller_sync"."teller_bank_account_id" = ?`, tellerBankAccountId).
		Order(`timestamp DESC`).
		Limit(1).
		Select(&sync)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to retrieve latest teller sync")
	}

	return &sync, nil
}
