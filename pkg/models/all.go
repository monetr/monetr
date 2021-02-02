package models

var (
	AllModels = []interface{}{
		&Login{},
		&EmailVerification{},
		&PhoneVerification{},
		&Account{},
		&User{},
		&Link{},
		&BankAccount{},
		&FundingSchedule{},
		&Expense{},
		&Transaction{},
	}
)
