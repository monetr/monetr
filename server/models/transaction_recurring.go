package models

import "time"

type WindowType string

const (
	FirstAndFifteenthWindowType WindowType = "FirstAndFifteenth"
	FifteenthAndLastWindowType  WindowType = "FifteenthAndLast"
	WeeklyWindowType            WindowType = "Weekly"
	BiWeeklyWindowType          WindowType = "BiWeekly"
	MonthlyWindowType           WindowType = "Monthly"
	BiMonthlyWindowType         WindowType = "BiMonthly"
	QuarterlyWindowType         WindowType = "Quarterly"
	SemiYearlyWindowType        WindowType = "SemiYearly"
	YearlyWindowType            WindowType = "Yearly"
)

type TransactionRecurring struct {
	tableName string `pg:"transaction_recurring"`

	TransactionRecurringId string            `json:"transactionRecurringId" pg:"transaction_recurring_id,notnull,pk"`
	AccountId              uint64            `json:"-" pg:"account_id,notnull,type:'bigint'"`
	Account                *Account          `json:"-" pg:"rel:has-one"`
	BankAccountId          uint64            `json:"bankAccountId" pg:"bank_account_id,notnull,type:'bigint'"`
	BankAccount            *BankAccount      `json:"-" pg:"rel:has-one"`
	Name                   string            `json:"name" pg:"name,notnull"`
	Window                 WindowType        `json:"windowType" pg:"window_type,notnull"`
	RuleSet                *RuleSet          `json:"ruleset" pg:"ruleset,notnull,type:'text'"`
	First                  time.Time         `json:"first" pg:"first,notnull"`
	Last                   time.Time         `json:"last" pg:"last,notnull"`
	Next                   time.Time         `json:"next" pg:"next,notnull"`
	Ended                  bool              `json:"ended" pg:"ended,notnull"`
	Confidence             float32           `json:"confidence" pg:"confidence,notnull"` // TODO What type should this be in postgres?
	Amounts                map[int64]int     `json:"amounts" pg:"amounts,notnull"`       // TODO This will be a JSONB or hashmap col.
	LastAmount             int64             `json:"lastAmount" pg:"last_amount,notnull"`
	Members                []ID[Transaction] `json:"members" pg:"members,notnull,type:'bigint[]'"`
	CreatedAt              time.Time         `json:"createdAt" pg:"created_at,notnull,default:now()"`
}
