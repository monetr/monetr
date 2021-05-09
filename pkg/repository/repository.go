package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/pkg/errors"
)

type Repository interface {
	AccountId() uint64
	UserId() uint64

	AddExpenseToTransaction(transaction *models.Transaction, spending *models.Spending) error
	CreateBankAccounts(bankAccounts ...models.BankAccount) error
	CreateFundingSchedule(fundingSchedule *models.FundingSchedule) error
	CreateLink(link *models.Link) error
	CreatePlaidLink(link *models.PlaidLink) error
	CreateSpending(expense *models.Spending) error
	CreateTransaction(bankAccountId uint64, transaction *models.Transaction) error
	DeleteSpending(ctx context.Context, bankAccountId, spendingId uint64) error
	DeleteTransaction(bankAccountId, transactionId uint64) error
	GetAccount() (*models.Account, error)
	GetBalances(ctx context.Context, bankAccountId uint64) (*Balances, error)
	GetBankAccount(bankAccountId uint64) (*models.BankAccount, error)
	GetBankAccounts() ([]models.BankAccount, error)
	GetBankAccountsByLinkId(linkId uint64) ([]models.BankAccount, error)
	GetFundingSchedule(bankAccountId, fundingScheduleId uint64) (*models.FundingSchedule, error)
	GetFundingSchedules(bankAccountId uint64) ([]models.FundingSchedule, error)
	GetFundingStats(ctx context.Context, bankAccountId uint64) (*FundingStats, error)
	GetIsSetup() (bool, error)
	GetJob(jobId string) (models.Job, error)
	GetLink(ctx context.Context, linkId uint64) (*models.Link, error)
	GetLinkIsManual(linkId uint64) (bool, error)
	GetLinkIsManualByBankAccountId(bankAccountId uint64) (bool, error)
	GetLinks() ([]models.Link, error)
	GetMe() (*models.User, error)
	GetPendingTransactionsForBankAccount(bankAccountId uint64) ([]models.Transaction, error)
	GetSpending(ctx context.Context, bankAccountId uint64) ([]models.Spending, error)
	GetSpendingByFundingSchedule(bankAccountId, fundingScheduleId uint64) ([]models.Spending, error)
	GetSpendingById(bankAccountId, expenseId uint64) (*models.Spending, error)
	GetTransaction(bankAccountId, transactionId uint64) (*models.Transaction, error)
	GetTransactions(bankAccountId uint64, limit, offset int) ([]models.Transaction, error)
	GetTransactionsByPlaidId(linkId uint64, plaidTransactionIds []string) (map[string]models.Transaction, error)
	GetTransactionsByPlaidTransactionId(linkId uint64, plaidTransactionIds []string) ([]models.Transaction, error)
	InsertTransactions(transactions []models.Transaction) error
	ProcessTransactionSpentFrom(bankAccountId uint64, input, existing *models.Transaction) (updatedExpenses []models.Spending, _ error)
	UpdateBankAccounts(accounts []models.BankAccount) error
	UpdateExpenses(bankAccountId uint64, updates []models.Spending) error
	UpdateLink(link *models.Link) error
	UpdateNextFundingScheduleDate(fundingScheduleId uint64, nextOccurrence time.Time) error
	UpdateTransaction(bankAccountId uint64, transaction *models.Transaction) error

	// UpdateTransactions is unique in that it REQUIRES that all data on each transaction object be populated. It is
	// doing a bulk update, so if data is missing it has the potential to overwrite a transaction incorrectly.
	UpdateTransactions(ctx context.Context, transactions []*models.Transaction) error
}

type UnauthenticatedRepository interface {
	CreateLogin(email, hashedPassword string, firstName, lastName string, isEnabled bool) (*models.Login, error)
	CreateAccount(timezone *time.Location) (*models.Account, error)
	CreateUser(loginId, accountId uint64, user *models.User) error

	// VerifyRegistration takes a registrationId and will finalize the registration record. If the registration has
	// already been completed an error is returned.
	VerifyRegistration(registrationId string) (*models.User, error)
	GetLinksForItem(itemId string) (*models.Link, error)
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

func NewUnauthenticatedRepository(txn *pg.Tx) UnauthenticatedRepository {
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

func (r *repositoryBase) GetMe() (*models.User, error) {
	var user models.User
	err := r.txn.Model(&user).
		Relation("Login").
		Relation("Account").
		Where(`"user"."user_id" = ? AND "user"."account_id" = ?`, r.userId, r.accountId).
		Limit(1).
		Select(&user)
	switch err {
	case pg.ErrNoRows:
		return nil, errors.Errorf("user does not exist")
	case nil:
		break
	default:
		return nil, errors.Wrapf(err, "failed to retrieve user")
	}

	return &user, nil
}

func (r *repositoryBase) GetIsSetup() (bool, error) {
	return r.txn.Model(&models.Link{}).
		Where(`"link"."account_id" = ?`, r.accountId).
		Exists()
}

func (r *repositoryBase) GetBankAccounts() ([]models.BankAccount, error) {
	var result []models.BankAccount
	err := r.txn.Model(&result).
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Select(&result)
	return result, errors.Wrap(err, "failed to retrieve bank accounts")
}
