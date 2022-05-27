package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

type BaseRepository interface {
	AccountId() uint64

	AddExpenseToTransaction(ctx context.Context, transaction *models.Transaction, spending *models.Spending) error
	CreateBankAccounts(ctx context.Context, bankAccounts ...models.BankAccount) error
	CreateFundingSchedule(ctx context.Context, fundingSchedule *models.FundingSchedule) error
	CreateLink(ctx context.Context, link *models.Link) error
	CreatePlaidLink(ctx context.Context, link *models.PlaidLink) error
	CreateSpending(ctx context.Context, expense *models.Spending) error
	CreateTransaction(ctx context.Context, bankAccountId uint64, transaction *models.Transaction) error
	// DeleteAccount removes all of the records from the database related to the current account. This action cannot be
	// undone. Any Plaid links should be removed BEFORE calling this function.
	DeleteAccount(ctx context.Context) error
	DeleteFundingSchedule(ctx context.Context, bankAccountId, fundingScheduleId uint64) error
	DeletePlaidLink(ctx context.Context, plaidLinkId uint64) error
	DeleteSpending(ctx context.Context, bankAccountId, spendingId uint64) error
	DeleteTransaction(ctx context.Context, bankAccountId, transactionId uint64) error
	GetAccount(ctx context.Context) (*models.Account, error)
	GetBalances(ctx context.Context, bankAccountId uint64) (*Balances, error)
	GetBankAccount(ctx context.Context, bankAccountId uint64) (*models.BankAccount, error)
	GetBankAccounts(ctx context.Context) ([]models.BankAccount, error)
	GetBankAccountsByLinkId(ctx context.Context, linkId uint64) ([]models.BankAccount, error)
	GetFundingSchedule(ctx context.Context, bankAccountId, fundingScheduleId uint64) (*models.FundingSchedule, error)
	GetFundingSchedules(ctx context.Context, bankAccountId uint64) ([]models.FundingSchedule, error)
	GetFundingStats(ctx context.Context, bankAccountId uint64) ([]FundingStats, error)
	GetIsSetup(ctx context.Context) (bool, error)
	GetLink(ctx context.Context, linkId uint64) (*models.Link, error)
	GetLinkIsManual(ctx context.Context, linkId uint64) (bool, error)
	GetLinkIsManualByBankAccountId(ctx context.Context, bankAccountId uint64) (bool, error)
	GetLinks(ctx context.Context) ([]models.Link, error)
	GetNumberOfPlaidLinks(ctx context.Context) (int, error)
	GetPendingTransactionsForBankAccount(ctx context.Context, bankAccountId uint64) ([]models.Transaction, error)
	GetSpending(ctx context.Context, bankAccountId uint64) ([]models.Spending, error)
	GetSpendingByFundingSchedule(ctx context.Context, bankAccountId, fundingScheduleId uint64) ([]models.Spending, error)
	GetSpendingById(ctx context.Context, bankAccountId, expenseId uint64) (*models.Spending, error)
	GetSpendingExists(ctx context.Context, bankAccountId, spendingId uint64) (bool, error)
	GetTransaction(ctx context.Context, bankAccountId, transactionId uint64) (*models.Transaction, error)
	GetTransactions(ctx context.Context, bankAccountId uint64, limit, offset int) ([]models.Transaction, error)
	// GetRecentDepositTransactions will return all deposit transactions for the specified bank account within the past
	// 24 hours.
	GetRecentDepositTransactions(ctx context.Context, bankAccountId uint64) ([]models.Transaction, error)
	GetTransactionsByPlaidId(ctx context.Context, linkId uint64, plaidTransactionIds []string) (map[string]models.Transaction, error)
	GetTransactionsByPlaidTransactionId(ctx context.Context, linkId uint64, plaidTransactionIds []string) ([]models.Transaction, error)
	GetTransactionsForSpending(ctx context.Context, bankAccountId, spendingId uint64, limit, offset int) ([]models.Transaction, error)
	InsertTransactions(ctx context.Context, transactions []models.Transaction) error
	ProcessTransactionSpentFrom(ctx context.Context, bankAccountId uint64, input, existing *models.Transaction) (updatedExpenses []models.Spending, _ error)
	UpdateBankAccounts(ctx context.Context, accounts []models.BankAccount) error
	UpdateSpending(ctx context.Context, bankAccountId uint64, updates []models.Spending) error
	UpdateLink(ctx context.Context, link *models.Link) error
	UpdateNextFundingScheduleDate(ctx context.Context, fundingScheduleId uint64, nextOccurrence time.Time) error
	UpdatePlaidLink(ctx context.Context, plaidLink *models.PlaidLink) error
	UpdateTransaction(ctx context.Context, bankAccountId uint64, transaction *models.Transaction) error

	// UpdateTransactions is unique in that it REQUIRES that all data on each transaction object be populated. It is
	// doing a bulk update, so if data is missing it has the potential to overwrite a transaction incorrectly.
	UpdateTransactions(ctx context.Context, transactions []*models.Transaction) error
}

type Repository interface {
	BaseRepository
	UserId() uint64

	GetMe(ctx context.Context) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

type UnauthenticatedRepository interface {
	CreateLogin(ctx context.Context, email, hashedPassword string, firstName, lastName string) (*models.Login, error)
	CreateAccountV2(ctx context.Context, account *models.Account) error
	CreateUser(ctx context.Context, loginId, accountId uint64, user *models.User) error
	GetLoginForEmail(ctx context.Context, emailAddress string) (*models.Login, error)
	ResetPassword(ctx context.Context, loginId uint64, hashedPassword string) error
	GetLinksForItem(ctx context.Context, itemId string) (*models.Link, error)
	ValidateBetaCode(ctx context.Context, betaCode string) (*models.Beta, error)
	UseBetaCode(ctx context.Context, betaId, usedBy uint64) error
}

func NewRepositoryFromSession(userId, accountId uint64, database pg.DBI) Repository {
	return &repositoryBase{
		userId:    userId,
		accountId: accountId,
		txn:       database,
	}
}

func NewUnauthenticatedRepository(txn pg.DBI) UnauthenticatedRepository {
	return &unauthenticatedRepo{
		txn: txn,
	}
}

var (
	_ Repository = &repositoryBase{}
)

func (r *repositoryBase) UserId() uint64 {
	return r.userId
}

func (r *repositoryBase) AccountId() uint64 {
	return r.accountId
}

func (r *repositoryBase) AccountIdStr() string {
	return strconv.FormatUint(r.AccountId(), 10)
}

func (r *repositoryBase) GetMe(ctx context.Context) (*models.User, error) {
	span := sentry.StartSpan(ctx, "GetMe")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
		"userId":    r.UserId(),
	}

	var user models.User
	err := r.txn.ModelContext(span.Context(), &user).
		Relation("Login").
		Relation("Account").
		Where(`"user"."user_id" = ? AND "user"."account_id" = ?`, r.userId, r.accountId).
		Limit(1).
		Select(&user)
	switch err {
	case pg.ErrNoRows:
		span.Status = sentry.SpanStatusNotFound
		return nil, errors.Errorf("user does not exist")
	case nil:
		break
	default:
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrapf(err, "failed to retrieve user")
	}

	span.Status = sentry.SpanStatusOK

	return &user, nil
}

func (r *repositoryBase) GetIsSetup(ctx context.Context) (bool, error) {
	span := sentry.StartSpan(ctx, "GetIsSetup")
	defer span.Finish()

	return r.txn.ModelContext(span.Context(), &models.Link{}).
		Where(`"link"."account_id" = ?`, r.accountId).
		Exists()
}
