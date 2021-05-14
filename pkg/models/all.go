package models

var (
	AllModels = []interface{}{
		&Login{},
		&Account{},
		&User{},
		&Job{},
		&PlaidLink{},
		&Link{},
		&BankAccount{},
		&FundingSchedule{},
		&Spending{},
		&Transaction{},
		&Subscription{},
	}

	// This silences any warnings about the tableName field not being used. It's used via reflection in our ORM to
	// query and generate schemas/SQL.
	_ = Account{}.tableName
	_ = BankAccount{}.tableName
	_ = FundingSchedule{}.tableName
	_ = Job{}.tableName
	_ = Link{}.tableName
	_ = Login{}.tableName
	_ = PlaidLink{}.tableName
	_ = Spending{}.tableName
	_ = Transaction{}.tableName
	_ = User{}.tableName
	_ = Subscription{}.tableName
)
