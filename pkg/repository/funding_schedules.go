package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetFundingSchedules(bankAccountId uint64) ([]models.FundingSchedule, error) {
	var result []models.FundingSchedule
	err := r.txn.Model(&result).
		Where(`"funding_schedule"."account_id" = ?`, r.AccountId()).
		Where(`"funding_schedule"."bank_account_id" = ?`, bankAccountId).
		Select(&result)
	return result, errors.Wrap(err, "failed to retrieve funding schedules")
}

func (r *repositoryBase) CreateFundingSchedule(fundingSchedule *models.FundingSchedule) error {
	return nil
}