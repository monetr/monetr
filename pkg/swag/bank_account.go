package swag

import (
	"github.com/monetr/monetr/pkg/models"
	"time"
)

type AlwaysBankAccount struct {
	// The balance available in the account represented as whole cents. This is typically the current balance minus the
	// total value of all pending transactions. This value is not calculated in the API and is retrieved from Plaid or
	// maintained manually for manual links.
	AvailableBalance int64 `json:"availableBalance" example:"102356"`
	// The current balance in the account as whole cents without taking into consideration any pending transactions.
	CurrentBalance int64 `json:"currentBalance" example:"102400"`
	// Last 4 digits of the bank account's account number. We do not store the full bank account number or any other
	// sensitive account information.
	Mask string `json:"mask" pg:"mask" example:"9876"`
	// Name of the account, this is different than the `originalName`. This field can be changed later on while the
	// `originalName` field cannot be changed once the account is created.
	Name string `json:"name,omitempty" example:"Checking Account"`
}

type UpdateBankAccountRequest struct {
	AlwaysBankAccount
}

type CreateBankAccountRequest struct {
	AlwaysBankAccount
	// The numeric Id of the Link this bank account is associated with, if the link is manual then bank bank accounts
	// can be created manually via the API. If the Link is associated with Plaid though then bank accounts can only be
	// created through the Plaid interface. At the time of writing this there is not a way to add or remove a bank
	// account from an existing Plaid Link.
	LinkId uint64 `json:"linkId" example:"2345" validate:"required"`
	// Account Type can be; depository, credit, loan, investment or other. At the time of writing this the application
	// will only support depository. Other types may be supported in the future.
	Type models.BankAccountType `json:"accountType" example:"depository" validate:"required"`
	// Sub Type can have numerous values, but given that the application currently only supports depository the most
	// common values you will see or use are; checking and savings. Other supported types (albeit untested) are; hsa,
	// cd, money market, paypal, prepaid, cash management and ebt.
	// More information on these can be found here: https://plaid.com/docs/api/accounts/#account-type-schema
	SubType models.BankAccountSubType `json:"accountSubType" example:"checking" validate:"required"`
	// The original name of the bank account from when it was created. This name cannot be changed after the bank
	// account is created. This is primarily due to bank account's coming from a 3rd party provider like Plaid. But to
	// reduce the amount of logic in the application the same rule applies for manual links as well.
	PlaidName string `json:"originalName" example:"Checking Account #1"`
	// Official name is only used with bank accounts coming from Plaid. It is another name that Plaid uses for an
	// account.
	PlaidOfficialName string `json:"officialName" example:"US Bank - Checking Account"`
}

type BankAccountResponse struct {
	// Bank Account ID uniquely identifies this bank account within monetr.
	BankAccountId uint64 `json:"bankAccountId" example:"85473" validate:"required"`

	CreateBankAccountRequest

	// Last Updated is used to keep track of changes to this specific bank account. It is updated when the balances that
	// are provided by an external party (like Plaid) are changed. It is not necessarily changed when the Safe-To-Spend
	// balance changes or when other fields on the bank account change.
	LastUpdated time.Time `json:"lastUpdated" example:"2021-04-15T00:00:00-05:00"`
}
