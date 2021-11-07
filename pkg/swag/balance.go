package swag

type BalanceResponse struct {
	// The bank account the balances are for. Balances are only per bank account, and not currently calculated at a link
	// or global level.
	BankAccountId uint64 `json:"bankAccountId" example:"1234"`
	// The current balance of the account in cents. This typically excludes pending transaction values.
	Current int64 `json:"current" example:"124396"`
	// The available balance of the account, usually the current balance minus any pending transactions.
	Available int64 `json:"available" example:"124000"`
	// The amount left over in the bank account after all expense and goal allocations have been subtracted from the
	// available balance.
	Safe int64 `json:"safe" example:"12350"`
	// The amount allocated to expense spending objects.
	Expenses int64 `json:"expenses" example:"100000"`
	// The amount allocated to goal spending objects.
	Goals int64 `json:"goals" example:"11650"`
}
