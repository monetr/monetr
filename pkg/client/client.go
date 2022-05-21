package client

import (
	"context"

	"github.com/monetr/monetr/pkg/models"
)

type MonetrClient interface {
	GetTransactions(ctx context.Context, bankAccountId uint64, count, offset int64) ([]models.Transaction, error)
	GetSpending(ctx context.Context, bankAccountId uint64) ([]models.Spending, error)
	GetFundingSchedules(ctx context.Context, bankAccountId uint64) ([]models.FundingSchedule, error)
	GetBankAccounts(ctx context.Context) ([]models.BankAccount, error)
	GetLinks(ctx context.Context) ([]models.Link, error)
	GetMe(ctx context.Context) (*models.User, error)
}

