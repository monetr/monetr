package models

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/pkg/errors"
)

type Transaction struct {
	tableName string `pg:"transactions"`

	TransactionId             uint64       `json:"transactionId" pg:"transaction_id,notnull,pk,type:'bigserial'"`
	AccountId                 uint64       `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account                   *Account     `json:"-" pg:"rel:has-one"`
	BankAccountId             uint64       `json:"bankAccountId" pg:"bank_account_id,notnull,pk,on_delete:CASCADE,type:'bigint',unique:per_bank_account"`
	BankAccount               *BankAccount `json:"-" pg:"rel:has-one"`
	PlaidTransactionId        string       `json:"-" pg:"plaid_transaction_id,unique:per_bank_account"`
	PendingPlaidTransactionId *string      `json:"-" pg:"pending_plaid_transaction_id"`
	Amount                    int64        `json:"amount" pg:"amount,notnull,use_zero"`
	SpendingId                *uint64      `json:"spendingId" pg:"spending_id,on_delete:SET NULL"`
	Spending                  *Spending    `json:"spending,omitempty" pg:"rel:has-one"`
	// SpendingAmount is the amount deducted from the expense this transaction was spent from. This is used when a
	// transaction is more than the expense currently has allocated. If the transaction were to be deleted or changed we
	// want to make sure we return the correct amount to the expense.
	SpendingAmount       *int64     `json:"spendingAmount,omitempty" pg:"spending_amount,use_zero"`
	Categories           []string   `json:"categories" pg:"categories,type:'text[]'"`
	OriginalCategories   []string   `json:"originalCategories" pg:"original_categories,type:'text[]'"`
	Date                 time.Time  `json:"date" pg:"date,notnull"`
	AuthorizedDate       *time.Time `json:"authorizedDate" pg:"authorized_date"`
	Name                 string     `json:"name,omitempty" pg:"name"`
	CustomName           *string    `json:"customName" pg:"custom_name"`
	OriginalName         string     `json:"originalName" pg:"original_name,notnull"`
	MerchantName         string     `json:"merchantName,omitempty" pg:"merchant_name"`
	OriginalMerchantName string     `json:"originalMerchantName" pg:"original_merchant_name"`
	Currency             string     `json:"currency" pg:"currency,notnull"`
	IsPending            bool       `json:"isPending" pg:"is_pending,notnull,use_zero"`
	CreatedAt            time.Time  `json:"createdAt" pg:"created_at,notnull,default:now()"`
	DeletedAt            *time.Time `json:"deletedAt" pg:"deleted_at"`
}

func (t Transaction) IsAddition() bool {
	return t.Amount < 0 // Deposits will show as negative amounts.
}

// AddSpendingToTransaction will take the provided spending object and deduct as much as possible from this transaction
// from that spending object. It does not change the spendingId on the transaction, it simply performs the deductions.
func (t *Transaction) AddSpendingToTransaction(ctx context.Context, spending *Spending, account *Account) error {
	span := sentry.StartSpan(ctx, "AddSpendingToTransaction")
	defer span.Finish()

	var allocationAmount int64
	// If the amount allocated to the spending we are adding to the transaction is less than the amount of the
	// transaction then we can only do a partial allocation.
	if spending.CurrentAmount < t.Amount {
		allocationAmount = spending.CurrentAmount
	} else {
		// Otherwise, we will allocate the entire transaction amount from the spending.
		allocationAmount = t.Amount
	}

	// Subtract the amount we are taking from the spending from it's current amount.
	spending.CurrentAmount -= allocationAmount

	switch spending.SpendingType {
	case SpendingTypeExpense:
	// We don't need to do anything special if it's an expense, at least not right now.
	case SpendingTypeGoal:
		// Goals also keep track of how much has been spent, so increment the used amount.
		spending.UsedAmount += allocationAmount
	}

	// Keep track of how much we took from the spending in case things change later.
	t.SpendingAmount = &allocationAmount

	// Now that we have deducted the amount we need from the spending we need to recalculate it's next contribution.
	if err := spending.CalculateNextContribution(
		span.Context(),
		account.Timezone,
		spending.FundingSchedule,
		time.Now(),
	); err != nil {
		return errors.Wrap(err, "failed to calculate next contribution for new transaction expense")
	}

	return nil
}

func AddSpendingToTransaction(
	ctx context.Context,
	transaction Transaction,
	spending Spending,
	timezone *time.Location,
	now time.Time,
) (amount int64, result Spending) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var allocationAmount int64
	// If the amount allocated to the spending we are adding to the transaction is less than the amount of the
	// transaction then we can only do a partial allocation.
	if spending.CurrentAmount < transaction.Amount {
		allocationAmount = spending.CurrentAmount
	} else {
		// Otherwise, we will allocate the entire transaction amount from the spending.
		allocationAmount = transaction.Amount
	}

	// Subtract the amount we are taking from the spending from it's current amount.
	spending.CurrentAmount -= allocationAmount

	switch spending.SpendingType {
	case SpendingTypeExpense:
	// We don't need to do anything special if it's an expense, at least not right now.
	case SpendingTypeGoal:
		// Goals also keep track of how much has been spent, so increment the used amount.
		spending.UsedAmount += allocationAmount
	}

	// Keep track of how much we took from the spending in case things change later.
	return allocationAmount, CalculateNextContribution(
		span.Context(),
		spending,
		*spending.FundingSchedule,
		timezone,
		now,
	)
}

func ProcessSpentFrom(
	ctx context.Context,
	input, currentTransaction Transaction,
	inputSpend, currentSpend *Spending,
	now time.Time,
	timezone *time.Location,
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

	var existingSpendingId uint64
	if currentSpend != nil {
		existingSpendingId = currentSpend.SpendingId
	}

	var newSpendingId uint64
	if inputSpend != nil {
		newSpendingId = inputSpend.SpendingId
	}

	var expensePlan int

	switch {
	case existingSpendingId == 0 && newSpendingId > 0:
		// Spending is being added to the transaction.
		expensePlan = AddExpense
	case existingSpendingId != 0 && newSpendingId != existingSpendingId && newSpendingId > 0:
		// Spending is being changed from one expense to another.
		expensePlan = ChangeExpense
	case existingSpendingId != 0 && newSpendingId == 0:
		// Spending is being removed from the transaction.
		expensePlan = RemoveExpense
	default:
		// TODO Handle transaction amount changes with expenses.
		return
	}

	updatedSpending = make([]Spending, 0)

	switch expensePlan {
	case ChangeExpense, RemoveExpense:
		// If the transaction already has an expense then it should have an expense amount. If this is missing then
		// something is wrong.
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

		// Now that we have added that money back to the expense we need to calculate the expense's next contribution.
		updatedSpending = append(updatedSpending, CalculateNextContribution(
			span.Context(),
			*currentSpend,
			*currentSpend.FundingSchedule,
			timezone,
			now,
		))

		// If we are only removing the expense then we are done with this part.
		if expensePlan == RemoveExpense {
			break
		}

		// If we are changing the expense though then we want to fallthrough to handle the processing of the new
		// expense.
		fallthrough
	case AddExpense:

		amountSpent, updatedNewSpend := AddSpendingToTransaction(
			span.Context(),
			input,
			*inputSpend,
			timezone,
			now,
		)

		// Then take all the fields that have changed and throw them in our list of things to update.
		updatedSpending = append(updatedSpending, updatedNewSpend)
		updatedTransaction.SpendingAmount = &amountSpent
	}

	return updatedTransaction, updatedSpending
}
