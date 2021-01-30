package models

import (
	"time"
)

type FundingSchedule struct {
	tableName string `sql:"funding_schedules"`

	FundingScheduleId uint64       `json:"fundingScheduleId" sql:"funding_schedule_id,notnull,pk,type:'bigserial'"`
	AccountId         uint64       `json:"-" sql:"account_id,notnull,pk,on_delete:CASCADE"`
	Account           *Account     `json:"-" sql:"rel:has-one"`
	BankAccountId     uint64       `json:"bankAccountId" sql:"bank_account_id,notnull,pk,on_delete:CASCADE,unique:per_bank"`
	BankAccount       *BankAccount `json:"bankAccount,omitempty" sql:"rel:has-one"`
	Name              string       `json:"name" sql:"name,notnull,unique:per_bank"`
	Description       string       `json:"description,omitempty" sql:"description,null"`
	Rule              *Rule        `json:"rule" json:"rule,notnull,type:'text'"`
	LastOccurrence    *time.Time   `json:"lastOccurrence" sql:"last_occurrence,type:'date'"`
	NextOccurrence    time.Time    `json:"nextOccurrence" sql:"next_occurrence,type:'date'"`
}
