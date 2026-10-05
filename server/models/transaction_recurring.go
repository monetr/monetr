package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Direction is whether money is leaving the account or coming into it.
type Direction string

const (
	// DebitDirection is money leaving the account, transactions with a positive
	// amount.
	DebitDirection Direction = "debit"
	// CreditDirection is money coming into the account, transactions with a
	// negative amount.
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
	CreatedAt              time.Time                `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	UpdatedAt              time.Time                `json:"updatedAt" bun:"updated_at,notnull,default:now(),nullzero"`

	Spending        *Spending        `json:"spending,omitempty" bun:"rel:has-one,join:transaction_recurring_id=transaction_recurring_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	FundingSchedule *FundingSchedule `json:"fundingSchedule,omitempty" bun:"rel:has-one,join:transaction_recurring_id=transaction_recurring_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
}

func (TransactionRecurring) IdentityPrefix() string {
	return "txrc"
}

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
