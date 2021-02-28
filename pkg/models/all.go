package models

var (
	AllModels = []interface{}{
		&Login{},
		&Registration{},
		&EmailVerification{},
		&PhoneVerification{},
		&Account{},
		&User{},
		&Job{},
		&PlaidLink{},
		&Link{},
		&BankAccount{},
		&FundingSchedule{},
		&Expense{},
		&Transaction{},
	}

	// This silences any warnings about the tableName field not being used. It's used via reflection in our ORM to
	// query and generate schemas/SQL.
	_ = Account{}.tableName
	_ = BankAccount{}.tableName
	_ = EmailVerification{}.tableName
	_ = Expense{}.tableName
	_ = FundingSchedule{}.tableName
	_ = Job{}.tableName
	_ = Link{}.tableName
	_ = Login{}.tableName
	_ = PhoneVerification{}.tableName
	_ = PlaidLink{}.tableName
	_ = Registration{}.tableName
	_ = Transaction{}.tableName
	_ = User{}.tableName
)
