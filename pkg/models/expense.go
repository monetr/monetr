package models

import (
	"time"
)

type Expense struct {
	tableName string `pg:"expenses"`

	ExpenseId              uint64           `json:"expenseId" pg:"expense_id,notnull,pk,type:'bigserial'"`
	AccountId              uint64           `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account                *Account         `json:"-" pg:"rel:has-one"`
	BankAccountId          uint64           `json:"bankAccountId" pg:"bank_account_id,notnull,pk,unique:per_bank,on_delete:CASCADE,type:'bigint'"`
	BankAccount            *BankAccount     `json:"bankAccount,omitempty" pg:"rel:has-one"`
	FundingScheduleId      *uint64          `json:"fundingScheduleId" pg:"funding_schedule_id,on_delete:SET NULL"`
	FundingSchedule        *FundingSchedule `json:"fundingSchedule,omitempty" pg:"rel:has-one"`
	Name                   string           `json:"name" pg:"name,notnull,unique:per_bank"`
	Description            string           `json:"description,omitempty" pg:"description"`
	TargetAmount           int64            `json:"targetAmount" pg:"target_amount,notnull,use_zero"`
	CurrentAmount          int64            `json:"currentAmount" pg:"current_amount,notnull,use_zero"`
	RecurrenceRule         *Rule            `json:"recurrenceRule" pg:"recurrence_rule,notnull,type:'text'"`
	LastRecurrence         *time.Time       `json:"lastRecurrence" pg:"last_recurrence,type:'date'"`
	NextRecurrence         time.Time        `json:"nextRecurrence" pg:"next_recurrence,notnull,type:'date'"`
	NextContributionAmount int64            `json:"nextContributionAmount" pg:"next_contribution_amount,notnull,use_zero"`
	IsBehind               bool             `json:"isBehind" pg:"is_behind,notnull,use_zero"`
}
