package models

import (
	"time"

	"github.com/uptrace/bun"
)

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
	bun.BaseModel `bun:"table:transaction_recurring,alias:transaction_recurring"`

	TransactionRecurringId string            `json:"transactionRecurringId" bun:"transaction_recurring_id,notnull,pk"`
	AccountId              uint64            `json:"-" bun:"account_id,notnull,type:bigint,nullzero"`
	Account                *Account          `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId          uint64            `json:"bankAccountId" bun:"bank_account_id,notnull,type:bigint,nullzero"`
	BankAccount            *BankAccount      `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	Name                   string            `json:"name" bun:"name,notnull,nullzero"`
	Window                 WindowType        `json:"windowType" bun:"window_type,notnull,nullzero"`
	RuleSet                *RuleSet          `json:"ruleset" bun:"ruleset,notnull,type:text"`
	First                  time.Time         `json:"first" bun:"first,notnull,nullzero"`
	Last                   time.Time         `json:"last" bun:"last,notnull,nullzero"`
	Next                   time.Time         `json:"next" bun:"next,notnull,nullzero"`
	Ended                  bool              `json:"ended" bun:"ended,notnull,nullzero"`
	Confidence             float32           `json:"confidence" bun:"confidence,notnull,nullzero"` // TODO What type should this be in postgres?
	Amounts                map[int64]int     `json:"amounts" bun:"amounts,notnull"`                // TODO This will be a JSONB or hashmap col.
	LastAmount             int64             `json:"lastAmount" bun:"last_amount,notnull,nullzero"`
	Members                []ID[Transaction] `json:"members" bun:"members,notnull,array,nullzero"`
	CreatedAt              time.Time         `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
}
