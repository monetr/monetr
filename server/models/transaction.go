package models

import (
	"context"
	"time"

	"log/slog"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/uptrace/bun"
)

type TransactionSource string

const (
	TransactionSourcePlaid     TransactionSource = "plaid"
	TransactionSourceUpload    TransactionSource = "upload"
	TransactionSourceManual    TransactionSource = "manual"
	TransactionSourceLunchFlow TransactionSource = "lunch_flow"
)

type Transaction struct {
	bun.BaseModel `bun:"table:transactions,alias:transaction"`

	TransactionId             ID[Transaction]           `json:"transactionId" bun:"transaction_id,notnull,pk"`
	AccountId                 ID[Account]               `json:"-" bun:"account_id,notnull,pk"`
	Account                   *Account                  `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	BankAccountId             ID[BankAccount]           `json:"bankAccountId" bun:"bank_account_id,notnull,pk,unique:per_bank_account"`
	BankAccount               *BankAccount              `json:"-" bun:"rel:belongs-to,join:bank_account_id=bank_account_id,join:account_id=account_id"`
	PlaidTransactionId        *ID[PlaidTransaction]     `json:"-" bun:"plaid_transaction_id"`
	PlaidTransaction          *PlaidTransaction         `json:"plaidTransaction" bun:"rel:belongs-to,join:plaid_transaction_id=plaid_transaction_id,join:account_id=account_id"`
	PendingPlaidTransactionId *ID[PlaidTransaction]     `json:"-" bun:"pending_plaid_transaction_id"`
	PendingPlaidTransaction   *PlaidTransaction         `json:"pendingPlaidTransaction" bun:"rel:belongs-to,join:pending_plaid_transaction_id=plaid_transaction_id,join:account_id=account_id"` // fk: is the prefix of the column we want to use to join on in a multikey join.
	LunchFlowTransactionId    *ID[LunchFlowTransaction] `json:"-" bun:"lunch_flow_transaction_id"`
	LunchFlowTransaction      *LunchFlowTransaction     `json:"lunchFlowTransaction,omitempty" bun:"rel:belongs-to,join:lunch_flow_transaction_id=lunch_flow_transaction_id,join:account_id=account_id"`
	Amount                    int64                     `json:"amount" bun:"amount,notnull"`
	SpendingId                *ID[Spending]             `json:"spendingId" bun:"spending_id"`
	Spending                  *Spending                 `json:"spending,omitempty" bun:"rel:belongs-to,join:spending_id=spending_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	// SpendingAmount is the amount deducted from the expense this transaction was
	// spent from. This is used when a transaction is more than the expense
	// currently has allocated. If the transaction were to be deleted or changed
	// we want to make sure we return the correct amount to the expense.
	SpendingAmount *int64 `json:"spendingAmount,omitempty" bun:"spending_amount"`
	// CreatedBySpendingId is set when the transaction was auto-created by the
	// ProcessSpending job for an expense with AutoCreateTransaction enabled. It
	// points at the spending that caused the transaction to be created and is
	// not user-mutable. This is distinct from SpendingId, which may be changed
	// by the user to re-allocate a transaction.
	CreatedBySpendingId *ID[Spending] `json:"createdBySpendingId" bun:"created_by_spending_id"`
	CreatedBySpending   *Spending     `json:"-" bun:"rel:belongs-to,join:created_by_spending_id=spending_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	// CreatedByFundingScheduleId is set when the transaction was auto-created
	// by the ProcessFundingSchedule job for a funding schedule with
	// AutoCreateTransaction enabled. It points at the funding schedule that
	// caused the transaction to be created and is not user-mutable.
	CreatedByFundingScheduleId *ID[FundingSchedule] `json:"createdByFundingScheduleId" bun:"created_by_funding_schedule_id"`
	CreatedByFundingSchedule   *FundingSchedule     `json:"-" bun:"rel:belongs-to,join:created_by_funding_schedule_id=funding_schedule_id,join:account_id=account_id,join:bank_account_id=bank_account_id"`
	Categories                 []string             `json:"categories" bun:"categories,array,nullzero"`
	Category                   *string              `json:"category" bun:"category"`
	Date                       time.Time            `json:"date" bun:"date,notnull,nullzero"`
	Name                       string               `json:"name,omitempty" bun:"name,nullzero"`
	OriginalName               string               `json:"originalName" bun:"original_name,notnull,nullzero"`
	MerchantName               string               `json:"merchantName,omitempty" bun:"merchant_name,nullzero"`
	OriginalMerchantName       string               `json:"originalMerchantName" bun:"original_merchant_name,nullzero"`
	IsPending                  bool                 `json:"isPending" bun:"is_pending,notnull"`
	UploadIdentifier           *string              `json:"uploadIdentifier" bun:"upload_identifier"`
	Source                     TransactionSource    `json:"source" bun:"source,nullzero"`
	CreatedAt                  time.Time            `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	DeletedAt                  *time.Time           `json:"deletedAt" bun:"deleted_at"`
}

func (Transaction) IdentityPrefix() string {
	return "txn"
}

var (
	_ bun.BeforeAppendModelHook = (*Transaction)(nil)
)

func (o *Transaction) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.TransactionId.IsZero() {
			o.TransactionId = NewID[Transaction]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}

func (t Transaction) IsAddition() bool {
	return t.Amount < 0 // Deposits will show as negative amounts.
}

// AddSpendingToTransaction will take the provided spending object and deduct as
// much as possible from this transaction from that spending object. It does not
// change the spendingId on the transaction, it simply performs the deductions.
func (t *Transaction) AddSpendingToTransaction(
	ctx context.Context,
	spending *Spending,
	timezone *time.Location,
	now time.Time,
	log *slog.Logger,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var allocationAmount int64
	// If the amount allocated to the spending we are adding to the transaction is
	// less than the amount of the transaction then we can only do a partial
	// allocation.
	if spending.CurrentAmount < t.Amount {
		allocationAmount = spending.CurrentAmount
	} else {
		// Otherwise, we will allocate the entire transaction amount from the
		// spending.
		allocationAmount = t.Amount
	}

	// Subtract the amount we are taking from the spending from it's current
	// amount.
	spending.CurrentAmount -= allocationAmount

	switch spending.SpendingType {
	case SpendingTypeExpense:
	// We don't need to do anything special if it's an expense, at least not right
	// now.
	case SpendingTypeGoal:
		// Goals also keep track of how much has been spent, so increment the used
		// amount.
		spending.UsedAmount += allocationAmount
	}

	// Keep track of how much we took from the spending in case things change
	// later.
	t.SpendingAmount = &allocationAmount

	// Now that we have deducted the amount we need from the spending we need to
	// recalculate it's next contribution.
	spending.CalculateNextContribution(
		span.Context(),
		timezone,
		spending.FundingSchedule,
		now,
		log,
	)

	return nil
}
