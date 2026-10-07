package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Direction string

// Debits are money leaving the account (positive amounts) and credits are money
// coming into the account (negative amounts).
const (
	DebitDirection  Direction = "debit"
	CreditDirection Direction = "credit"
)

type WindowType string

const (
	FirstAndFifteenthWindowType WindowType = "firstAndFifteenth"
	FifteenthAndLastWindowType  WindowType = "fifteenthAndLast"
	WeeklyWindowType            WindowType = "weekly"
	BiWeeklyWindowType          WindowType = "biweekly"
	MonthlyWindowType           WindowType = "monthly"
	BiMonthlyWindowType         WindowType = "bimonthly"
	QuarterlyWindowType         WindowType = "quarterly"
	SemiYearlyWindowType        WindowType = "semiyearly"
	YearlyWindowType            WindowType = "yearly"
)

type TransactionRecurring struct {
	bun.BaseModel `bun:"table:transaction_recurring,alias:transaction_recurring"`

	TransactionRecurringId ID[TransactionRecurring] `json:"transactionRecurringId" bun:"transaction_recurring_id,notnull,pk"`
	AccountId              ID[Account]              `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account                 `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId          ID[BankAccount]          `json:"bankAccountId" bun:"bank_account_id,notnull,pk"`
	BankAccount            *BankAccount             `json:"bankAccount,omitempty" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	TransactionClusterId   ID[TransactionCluster]   `json:"transactionClusterId" bun:"transaction_cluster_id,notnull"`
	TransactionCluster     *TransactionCluster      `json:"transactionCluster,omitempty" bun:"rel:belongs-to,join:transaction_cluster_id=transaction_cluster_id,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	SpendingId             *ID[Spending]            `json:"spendingId" bun:"spending_id"`
	FundingScheduleId      *ID[FundingSchedule]     `json:"fundingScheduleId" bun:"funding_schedule_id"`
	FundingSchedule        *FundingSchedule         `json:"fundingSchedule,omitempty" bun:"rel:belongs-to,join:funding_schedule_id=funding_schedule_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	Window                 WindowType               `json:"window" bun:"window_type,notnull,nullzero"`
	RuleSet                *RuleSet                 `json:"ruleset" bun:"ruleset,notnull,type:text"`
	First                  time.Time                `json:"first" bun:"first,notnull,nullzero"`
	Last                   time.Time                `json:"last" bun:"last,notnull,nullzero"`
	Next                   time.Time                `json:"next" bun:"next,notnull,nullzero"`
	Ended                  bool                     `json:"ended" bun:"ended,notnull,nullzero"`
	Confidence             float32                  `json:"confidence" bun:"confidence,notnull,nullzero"` // TODO What type should this be in postgres?
	Direction              Direction                `json:"direction" bun:"direction,notnull"`
	Amounts                map[int64]int            `json:"amounts" bun:"amounts,notnull"` // TODO This will be a JSONB or hashmap col.
	LastAmount             int64                    `json:"lastAmount" bun:"last_amount,notnull,nullzero"`
	// AutoMatched is true when the spending link was made by the matching job
	// instead of the user. It is cleared whenever the user changes the links.
	AutoMatched bool      `json:"autoMatched" bun:"auto_matched,notnull"`
	CreatedAt   time.Time `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	UpdatedAt   time.Time `json:"updatedAt" bun:"updated_at,notnull,default:now(),nullzero"`
}

func (TransactionRecurring) IdentityPrefix() string {
	return "txrc"
}

var (
	_ bun.BeforeAppendModelHook = (*TransactionRecurring)(nil)
)

func (o *TransactionRecurring) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.TransactionRecurringId.IsZero() {
			o.TransactionRecurringId = NewID[TransactionRecurring]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = now
		}
	}

	return nil
}
