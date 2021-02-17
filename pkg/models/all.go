package models

var (
	AllModels = []interface{}{
		&Login{},
		&Registration{},
		&EmailVerification{},
		&PhoneVerification{},
		&Account{},
		&User{},
		&PlaidLink{},
		&Link{},
		&BankAccount{},
		&FundingSchedule{},
		&Expense{},
		&Transaction{},
	}
)
