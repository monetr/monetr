package models

var (
	AllModels = []interface{}{
		&Login{},
		&Registration{},
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
