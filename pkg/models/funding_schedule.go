package models

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/util"
	"time"
)

type FundingSchedule struct {
	tableName string `pg:"funding_schedules"`

	FundingScheduleId uint64       `json:"fundingScheduleId" pg:"funding_schedule_id,notnull,pk,type:'bigserial'"`
	AccountId         uint64       `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account           *Account     `json:"-" pg:"rel:has-one"`
	BankAccountId     uint64       `json:"bankAccountId" pg:"bank_account_id,notnull,pk,on_delete:CASCADE,unique:per_bank,type:'bigint'"`
	BankAccount       *BankAccount `json:"bankAccount,omitempty" pg:"rel:has-one" swaggerignore:"true"`
	Name              string       `json:"name" pg:"name,notnull,unique:per_bank"`
	Description       string       `json:"description,omitempty" pg:"description"`
	Rule              *Rule        `json:"rule" pg:"rule,notnull,type:'text'" swaggertype:"string" example:"FREQ=MONTHLY;BYMONTHDAY=15,-1"`
	LastOccurrence    *time.Time   `json:"lastOccurrence" pg:"last_occurrence"`
	NextOccurrence    time.Time    `json:"nextOccurrence" pg:"next_occurrence,notnull"`
}

func (f *FundingSchedule) CalculateNextOccurrence(ctx context.Context, timezone *time.Location) bool {
	span := sentry.StartSpan(ctx, "CalculateNextOccurrence")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"fundingScheduleId": f.FundingScheduleId,
		"timezone":          timezone.String(),
	}

	now := time.Now()

	if now.Before(f.NextOccurrence) {
		crumbs.Debug(span.Context(), "Skipping processing funding schedule, it does not occur yet", map[string]interface{}{
			"fundingScheduleId": f.FundingScheduleId,
			"nextOccurrence":    f.NextOccurrence,
		})
		return false
	}

	nextFundingOccurrence := util.MidnightInLocal(f.Rule.After(now, false), timezone)

	current := f.NextOccurrence
	f.LastOccurrence = &current
	f.NextOccurrence = nextFundingOccurrence

	return true
}
