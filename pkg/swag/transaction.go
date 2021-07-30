package swag

import "time"

type AlwaysTransaction struct {
	// The amount of the transaction in cents. This is used when a transaction is "spent-from" a spending object.
	// **NOTE**: `amount` cannot be updated on transactions that were created from Plaid.
	Amount int64 `json:"amount" validate:"required" minimum:"1"`
	// The expense or goal's spending Id that this transaction was spent from. When this is updated the spending amount
	// will also be updated here. On the spending object the allocated amount will be subtracted from up to the amount
	// of this transaction. But the spending objects allocated amount will never be negative. If you have a transaction
	// that is $10.00 and you spend it from a spending object with only $8.00 allocated, then only $8.00 will be
	// subtracted from the spending object. Those $8.00 will be represented by the `spendingAmount` field here.
	SpendingId *uint64 `json:"spendingId" example:"54312" extensions:"x-nullable"`
	// Represents a path of categories that represents what type of spending this transaction was. For example:
	// `["Restaurants", "Fast Food"]`. A transaction could just have the category of `Restaurants`, but it can have a
	// child category of `Fast Food` as well. This field can be maintained directly by the end user. But is typically
	// generated when the transaction is created from Plaid.
	Categories []string `json:"categories" example:"Restaurants,Fast Food" extensions:"x-nullable"`
	// Date is the date the transaction was created. This date cannot change on this particular transaction Id, but if
	// the transaction is in a `Pending` state then when the transaction clears a new transaction can be created and
	// this transaction would be deleted. This can change the `date` field when this occurs.
	// **NOTE**: `date` cannot be updated on transactions that were created from Plaid.
	Date time.Time `json:"date" example:"2021-04-15T00:00:00-05:00"`
	// Authorized date comes from Plaid, but to my knowledge will not be populated in this API until we support UK
	// banks.
	// > This field is only populated for UK institutions. For institutions in other countries, will be null.
	// https://plaid.com/docs/api/products/#transactions-get-response-authorized-datetime_transactions
	// **NOTE**: `date` cannot be updated on transactions that were created from Plaid.
	AuthorizedDate *time.Time `json:"authorizedDate"`
	Name           string     `json:"name,omitempty"`
	MerchantName   string     `json:"merchantName,omitempty"`
	IsPending      bool       `json:"isPending" pg:"is_pending,notnull,use_zero"`
}

type UpdateTransactionRequest struct {
	AlwaysTransaction
}

type NewTransactionRequest struct {
	AlwaysTransaction

	// The Id of bank account that this transaction is associated with. A transaction can only be associated with a
	// single bank account. It is required when creating a new transaction.
	BankAccountId uint64 `json:"bankAccountId" example:"43872" validate:"required"`
	OriginalName  string `json:"originalName"`

	// Original merchant name is immutable. It can only be set when the transaction is created. This is used to preserve
	// some data about the original transaction and is primarily used by the Plaid integration. It is not required for
	// manually created transactions.
	OriginalMerchantName string `json:"originalMerchantName" example:"Uber"`

	OriginalCategories []string `json:"originalCategories" pg:"original_categories,type:'text[]'"`
}

type TransactionResponse struct {
	NewTransactionRequest
	UpdateTransactionRequest
	// The unique Id for the transaction within monetr. This is globally unique.
	TransactionId uint64 `json:"transactionId" example:"58732" validate:"required"`
	// SpendingAmount is the amount deducted from the expense this transaction was spent from. This is used when a
	// transaction is more than the expense currently has allocated. If the transaction were to be deleted or changed we
	// want to make sure we return the correct amount to the expense.
	SpendingAmount *int64 `json:"spendingAmount,omitempty" example:"800" extensions:"x-nullable"`

	CreatedAt time.Time `json:"createdAt"`
}
