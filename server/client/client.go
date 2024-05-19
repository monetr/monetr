package client

import (
	"context"

	. "github.com/monetr/monetr/server/models"
)

type MonetrClient interface {
	GetTransactions(ctx context.Context, bankAccountId ID[BankAccount], count, offset int64) ([]Transaction, error)
	GetSpending(ctx context.Context, bankAccountId ID[BankAccount]) ([]Spending, error)
	GetFundingSchedules(ctx context.Context, bankAccountId ID[BankAccount]) ([]FundingSchedule, error)
	GetBankAccounts(ctx context.Context) ([]BankAccount, error)
	GetLinks(ctx context.Context) ([]Link, error)
	GetMe(ctx context.Context) (*User, error)
}
