package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/sirupsen/logrus"
)

type TransactionSource string

const (
	TransactionSourcePlaid     TransactionSource = "plaid"
	TransactionSourceUpload    TransactionSource = "upload"
	TransactionSourceManual    TransactionSource = "manual"
	TransactionSourceLunchFlow TransactionSource = "lunchflow"
)

type Transaction struct {
	tableName string `pg:"transactions"`

	TransactionId             ID[Transaction]           `json:"transactionId" pg:"transaction_id,notnull,pk"`
	AccountId                 ID[Account]               `json:"-" pg:"account_id,notnull,pk"`
	Account                   *Account                  `json:"-" pg:"rel:has-one"`
	BankAccountId             ID[BankAccount]           `json:"bankAccountId" pg:"bank_account_id,notnull,pk,unique:per_bank_account"`
	BankAccount               *BankAccount              `json:"-" pg:"rel:has-one"`
	PlaidTransactionId        *ID[PlaidTransaction]     `json:"-" pg:"plaid_transaction_id"`
	PlaidTransaction          *PlaidTransaction         `json:"plaidTransaction" pg:"rel:has-one"`
	PendingPlaidTransactionId *ID[PlaidTransaction]     `json:"-" pg:"pending_plaid_transaction_id"`
	PendingPlaidTransaction   *PlaidTransaction         `json:"pendingPlaidTransaction" pg:"rel:has-one,fk:pending_"` // fk: is the prefix of the column we want to use to join on in a multikey join.
	LunchFlowTransactionId    *ID[LunchFlowTransaction] `json:"-" pg:"lunchflow_transaction_id"`
	LunchFlowTransaction      *LunchFlowTransaction     `json:"lunchFlowTransaction,omitempty" pg:"rel:has-one"`
	Amount                    int64                     `json:"amount" pg:"amount,notnull,use_zero"`
	SpendingId                *ID[Spending]             `json:"spendingId" pg:"spending_id,on_delete:SET NULL"`
	Spending                  *Spending                 `json:"spending,omitempty" pg:"rel:has-one"`
	// SpendingAmount is the amount deducted from the expense this transaction was
	// spent from. This is used when a transaction is more than the expense
	// currently has allocated. If the transaction were to be deleted or changed
	// we want to make sure we return the correct amount to the expense.
	SpendingAmount       *int64            `json:"spendingAmount,omitempty" pg:"spending_amount,use_zero"`
	Categories           []string          `json:"categories" pg:"categories,type:'text[]'"`
	Category             *string           `json:"category" pg:"category"`
	Date                 time.Time         `json:"date" pg:"date,notnull"`
	Name                 string            `json:"name,omitempty" pg:"name"`
	OriginalName         string            `json:"originalName" pg:"original_name,notnull"`
	MerchantName         string            `json:"merchantName,omitempty" pg:"merchant_name"`
	OriginalMerchantName string            `json:"originalMerchantName" pg:"original_merchant_name"`
	IsPending            bool              `json:"isPending" pg:"is_pending,notnull,use_zero"`
	UploadIdentifier     *string           `json:"uploadIdentifier" pg:"upload_identifier"`
	Source               TransactionSource `json:"source" pg:"source"`
	CreatedAt            time.Time         `json:"createdAt" pg:"created_at,notnull,default:now()"`
	DeletedAt            *time.Time        `json:"deletedAt" pg:"deleted_at"`
}

func (Transaction) IdentityPrefix() string {
	return "txn"
}

var (
	_ pg.BeforeInsertHook = (*Transaction)(nil)
)

func (o *Transaction) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.TransactionId.IsZero() {
		o.TransactionId = NewID[Transaction]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
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
	log *logrus.Entry,
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

func AddSpendingToTransaction(
	ctx context.Context,
	transaction Transaction,
	spending Spending,
	timezone *time.Location,
	now time.Time,
	log *logrus.Entry,
) (amount int64, result Spending) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var allocationAmount int64
	// If the amount allocated to the spending we are adding to the transaction is
	// less than the amount of the transaction then we can only do a partial
	// allocation.
	if spending.CurrentAmount < transaction.Amount {
		allocationAmount = spending.CurrentAmount
	} else {
		// Otherwise, we will allocate the entire transaction amount from the
		// spending.
		allocationAmount = transaction.Amount
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

	// Keep track of how much we took from the spending in case things change later.

	spending.CalculateNextContribution(
		span.Context(),
		timezone,
		spending.FundingSchedule,
		now,
		log,
	)

	return allocationAmount, spending
}

func ProcessSpentFrom(
	ctx context.Context,
	input, currentTransaction Transaction,
	inputSpend, currentSpend *Spending,
	now time.Time,
	timezone *time.Location,
	log *logrus.Entry,
) (updatedTransaction Transaction, updatedSpending []Spending) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	myownsanity.ASSERT_NOTNIL(timezone, "timezone is required to process spent from")

	updatedTransaction = input

	// Only a few different scenarios for what we can actually do.
	const (
		AddExpense = iota
		ChangeExpense
		RemoveExpense
	)

	var existingSpendingId ID[Spending]
	if currentSpend != nil {
		existingSpendingId = currentSpend.SpendingId
	}

	var newSpendingId ID[Spending]
	if inputSpend != nil {
		newSpendingId = inputSpend.SpendingId
	}

	var expensePlan int

	switch {
	case existingSpendingId.IsZero() && !newSpendingId.IsZero():
		// Spending is being added to the transaction.
		expensePlan = AddExpense
	case !existingSpendingId.IsZero() && newSpendingId != existingSpendingId && !newSpendingId.IsZero():
		// Spending is being changed from one expense to another.
		expensePlan = ChangeExpense
	case !existingSpendingId.IsZero() && newSpendingId.IsZero():
		// Spending is being removed from the transaction.
		expensePlan = RemoveExpense
	default:
		// TODO Handle transaction amount changes with expenses.
		return
	}

	updatedSpending = make([]Spending, 0)

	switch expensePlan {
	case ChangeExpense, RemoveExpense:
		// If the transaction already has an expense then it should have an expense
		// amount. If this is missing then something is wrong.
		myownsanity.ASSERT_NOTNIL(
			currentTransaction.SpendingAmount,
			"transaction spending amount can't be nil because it has been spent from something",
		)

		// Add the amount we took from the expense back to it.
		currentSpend.CurrentAmount += *currentTransaction.SpendingAmount

		switch currentSpend.SpendingType {
		case SpendingTypeExpense:
		// Nothing special for expenses.
		case SpendingTypeGoal:
			// Revert the amount used for the current spending object.
			currentSpend.UsedAmount -= *currentTransaction.SpendingAmount
		}

		updatedTransaction.SpendingAmount = nil

		myownsanity.ASSERT_NOTNIL(
			currentSpend.FundingSchedule,
			"current spend is missing the embedded funding schedule data",
		)

		{ // Clone the expense object so we don't modify things upstream
			current := *currentSpend
			// Calculate the next contribution
			current.CalculateNextContribution(
				span.Context(),
				timezone,
				current.FundingSchedule,
				now,
				log,
			)
			// Then add it to our list
			updatedSpending = append(updatedSpending, current)
		}

		// If we are only removing the expense then we are done with this part.
		if expensePlan == RemoveExpense {
			break
		}

		// If we are changing the expense though then we want to fallthrough to
		// handle the processing of the new expense.
		fallthrough
	case AddExpense:
		amountSpent, updatedNewSpend := AddSpendingToTransaction(
			span.Context(),
			input,
			*inputSpend,
			timezone,
			now,
			log,
		)

		// Then take all the fields that have changed and throw them in our list of
		// things to update.
		updatedSpending = append(updatedSpending, updatedNewSpend)
		updatedTransaction.SpendingAmount = &amountSpent
	}

	return updatedTransaction, updatedSpending
}

func (Transaction) CreateValidators() []*validation.KeyRules {
	return []*validation.KeyRules{
		validators.Name(validators.Require),
		validation.Key(
			"bankAccountId",
			validation.Required.Error("Must specify a bank account ID"),
			ValidID[BankAccount]().Error("Bank account ID specified is not valid"),
		).Required(validators.Require),
		validation.Key(
			"amount",
			// TODO Require that it is a number
			validation.Required.Error("Must specify a transaction amount"),
		).Required(validators.Require),
		validation.Key(
			"spendingId",
			ValidID[Spending]().Error("Spending ID specified is not valid"),
		).Required(validators.Optional),
		validation.Key(
			"spendingAmount",
			validation.Min(0).Error("Spending amount cannot be less than zero"),
		).Required(validators.Optional),
		validation.Key(
			"date",
			validation.Required.Error("Must specify a transaction date"),
			validation.Date(time.RFC3339).Error("Transaction date must be valid"),
		).Required(validators.Require),
		validation.Key(
			"isPending",
			validation.In(true, false).Error("Is pending must be true or false"),
		).Required(validators.Optional),
	}
}

func (Transaction) UpdateValidators() []*validation.KeyRules {
	return []*validation.KeyRules{
		validators.Name(validators.Optional),
		validation.Key(
			"amount",
			// TODO Require that it is a number
			validation.Required.Error("Must specify a transaction amount"),
		).Required(validators.Optional),
		validation.Key(
			"spendingId",
			ValidID[Spending]().Error("Spending ID specified is not valid"),
		).Required(validators.Optional),
		validation.Key(
			"spendingAmount",
			validation.Min(0).Error("Spending amount cannot be less than zero"),
		).Required(validators.Optional),
		validation.Key(
			"date",
			validation.Date(time.RFC3339).Error("Transaction date must be valid"),
		).Required(validators.Optional),
		validation.Key(
			"isPending",
			validation.In(true, false).Error("Is pending must be true or false"),
		).Required(validators.Optional),
	}
}
