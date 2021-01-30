package models

import (
	"time"
)

type Expense struct {
	tableName string `sql:"expenses"`

	ExpenseId              uint64           `json:"expenseId" sql:"expense_id,notnull,pk,type:'bigserial'"`
	AccountId              uint64           `json:"-" sql:"account_id,notnull,pk,on_delete:CASCADE"`
	Account                *Account         `json:"-" sql:"rel:has-one"`
	BankAccountId          uint64           `json:"bankAccountId" sql:"bank_account_id,notnull,pk,unique:per_bank,on_delete:CASCADE"`
	BankAccount            *BankAccount     `json:"bankAccount,omitempty" sql:"rel:has-one"`
	FundingScheduleId      *uint64          `json:"fundingScheduleId" sql:"funding_schedule_id,null,on_delete:SET NULL"`
	FundingSchedule        *FundingSchedule `json:"fundingSchedule,omitempty" sql:"rel:has-one"`
	Name                   string           `json:"name" sql:"name,notnull,unique:per_bank"`
	Description            string           `json:"description,omitempty" sql:"description,null"`
	TargetAmount           int64            `json:"targetAmount" sql:"target_amount,notnull,use_zero"`
	CurrentAmount          int64            `json:"currentAmount" sql:"current_amount,notnull,use_zero"`
	RecurrenceRule         *Rule            `json:"recurrenceRule" sql:"recurrence_rule,notnull,type:'text'"`
	LastRecurrence         *time.Time       `json:"lastRecurrence" sql:"last_recurrence,null,type:'date'"`
	NextRecurrence         time.Time        `json:"nextRecurrence" sql:"next_recurrence,notnull,type:'date'"`
	NextContributionAmount int64            `json:"nextContributionAmount" sql:"next_contribution_amount,notnull,use_zero"`
	IsBehind               bool             `json:"isBehind" sql:"is_behind,notnull,use_zero"`
}
