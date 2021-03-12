package models

import (
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
	LastOccurrence    *time.Time   `json:"lastOccurrence" pg:"last_occurrence,type:'date'"`
	NextOccurrence    time.Time    `json:"nextOccurrence" pg:"next_occurrence,type:'date'"`
}
