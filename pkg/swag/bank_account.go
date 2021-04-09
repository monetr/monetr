package swag

type BankAccountCreateRequest struct {
	// The numeric Id of the Link this bank account is associated with, if the link is manual then bank bank accounts
	// can be created manually via the API. If the Link is associated with Plaid though then bank accounts can only be
	// created through the Plaid interface. At the time of writing this there is not a way to add or remove a bank
	// account from an existing Plaid Link.
	LinkId           uint64   `json:"linkId" example:"2345"`
	// The balance available in the account represented as whole cents. This is typically the current balance minus the
	// total value of all pending transactions. This value is not calculated in the API and is retrieved from Plaid or
	// maintained manually for manual links.
	AvailableBalance int64    `json:"availableBalance" example:"102356"`
	// The current balance in the account as whole cents without taking into consideration any pending transactions.
	CurrentBalance    int64  `json:"currentBalance" example:"102400"`
	// Last 4 digits of the bank account's account number. We do not store the full bank account number or any other
	// sensitive account information.
	Mask              string `json:"mask" pg:"mask" example:"9876"`
	// Name of the account, this is different than the `originalName`. This field can be changed later on while the
	// `originalName` field cannot be changed once the account is created.
	Name              string `json:"name,omitempty" example:"Checking Account"`
	PlaidName         string `json:"originalName" example:"Checking Account #1"`
	PlaidOfficialName string `json:"officialName" example:"US Bank - Checking Account"`
	Type              string `json:"accountType" example:"depository"`
	SubType           string `json:"accountSubType" example:"checking"`
}
