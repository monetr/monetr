package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
)

type BaseRepository interface {
	AccountId() uint64

	CreatePlaidBankAccount(ctx context.Context, bankAccount *models.PlaidBankAccount) error
	UpdatePlaidBankAccount(ctx context.Context, bankAccount *models.PlaidBankAccount) error

	GetPlaidBankAccountsByLinkId(ctx context.Context, linkId uint64) ([]models.PlaidBankAccount, error)

	AddExpenseToTransaction(ctx context.Context, transaction *models.Transaction, spending *models.Spending) error
	CreateBankAccounts(ctx context.Context, bankAccounts ...*models.BankAccount) error
	CreateFundingSchedule(ctx context.Context, fundingSchedule *models.FundingSchedule) error
	CreateLink(ctx context.Context, link *models.Link) error
	CreatePlaidLink(ctx context.Context, link *models.PlaidLink) error
	CreateSpending(ctx context.Context, expense *models.Spending) error
	CreateTransaction(ctx context.Context, bankAccountId uint64, transaction *models.Transaction) error

	// CreatePlaidTransaction takes a Plaid transaction model and ensures the
	// account ID and the created at timestamp are properly set then stores the
	// transaction in the database.
	CreatePlaidTransaction(ctx context.Context, transaction *models.PlaidTransaction) error

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
	// GetBankAccountsWithPlaidByLinkId will return all the bank accounts
	// associated with the provided link ID that also have a Plaid bank account
	// associated with them.
	GetBankAccountsWithPlaidByLinkId(ctx context.Context, linkId uint64) ([]models.BankAccount, error)
	GetFundingSchedule(ctx context.Context, bankAccountId, fundingScheduleId uint64) (*models.FundingSchedule, error)
	GetFundingSchedules(ctx context.Context, bankAccountId uint64) ([]models.FundingSchedule, error)
	GetFundingStats(ctx context.Context, bankAccountId uint64) ([]FundingStats, error)
	GetIsSetup(ctx context.Context) (bool, error)
	GetLink(ctx context.Context, linkId uint64) (*models.Link, error)
	GetLinkIsManual(ctx context.Context, linkId uint64) (bool, error)
	GetLinkIsManualByBankAccountId(ctx context.Context, bankAccountId uint64) (bool, error)
	GetLinks(ctx context.Context) ([]models.Link, error)

	// Plaid syncing
	GetLastPlaidSync(ctx context.Context, linkId uint64) (*models.PlaidSync, error)
	RecordPlaidSync(ctx context.Context, plaidLinkId uint64, trigger, nextCursor string, added, modified, removed int) error

	GetNumberOfPlaidLinks(ctx context.Context) (int, error)
	GetSettings(ctx context.Context) (*models.Settings, error)
	GetSpending(ctx context.Context, bankAccountId uint64) ([]models.Spending, error)
	GetSpendingByFundingSchedule(ctx context.Context, bankAccountId, fundingScheduleId uint64) ([]models.Spending, error)
	GetSpendingById(ctx context.Context, bankAccountId, expenseId uint64) (*models.Spending, error)
	GetSpendingExists(ctx context.Context, bankAccountId, spendingId uint64) (bool, error)
	GetTransaction(ctx context.Context, bankAccountId, transactionId uint64) (*models.Transaction, error)
	GetTransactions(ctx context.Context, bankAccountId uint64, limit, offset int) ([]models.Transaction, error)
	// GetTransactionsAfter will return all of the transactions after the
	// specified date, if the specified date is null then all transactions for an
	// account is returned. This is intended to be used for partial syncing for
	// file uploads or teller.
	GetTransactionsAfter(ctx context.Context, bankAccountId uint64, after *time.Time) ([]models.Transaction, error)
	// GetPendingTransactions is the same as GetTransactions but will only return
	// transactions that are currently in a pending state. It will not return
	// transactions that have been deleted.
	GetPendingTransactions(ctx context.Context, bankAccountId uint64, limit, offset int) ([]models.Transaction, error)
	// GetRecentDepositTransactions will return all deposit transactions for the specified bank account within the past
	// 24 hours.
	GetRecentDepositTransactions(ctx context.Context, bankAccountId uint64) ([]models.Transaction, error)
	GetTransactionsByPlaidId(ctx context.Context, linkId uint64, plaidTransactionIds []string) (map[string]models.Transaction, error)
	GetTransactionsByPlaidTransactionId(ctx context.Context, linkId uint64, plaidTransactionIds []string) ([]models.Transaction, error)
	GetTransactionsForSpending(ctx context.Context, bankAccountId, spendingId uint64, limit, offset int) ([]models.Transaction, error)
	InsertTransactions(ctx context.Context, transactions []models.Transaction) error
	ProcessTransactionSpentFrom(ctx context.Context, bankAccountId uint64, input, existing *models.Transaction) (updatedExpenses []models.Spending, _ error)
	UpdateBankAccounts(ctx context.Context, accounts ...models.BankAccount) error
	UpdateSpending(ctx context.Context, bankAccountId uint64, updates []models.Spending) error
	UpdateLink(ctx context.Context, link *models.Link) error
	UpdateFundingSchedule(ctx context.Context, fundingSchedule *models.FundingSchedule) error
	UpdatePlaidLink(ctx context.Context, plaidLink *models.PlaidLink) error
	UpdateTransaction(ctx context.Context, bankAccountId uint64, transaction *models.Transaction) error

	// UpdateTransactions is unique in that it REQUIRES that all data on each transaction object be populated. It is
	// doing a bulk update, so if data is missing it has the potential to overwrite a transaction incorrectly.
	UpdateTransactions(ctx context.Context, transactions []*models.Transaction) error

	// WriteTransactionClusters will take the array of transaction clusters
	// provided and persist them to the trnasaction clusters table but will also
	// delete any existing transaction clusters for the bank account specified.
	// This is because clusters are meant to be regenerated each time new
	// transactions come in.
	WriteTransactionClusters(ctx context.Context, bankAccountId uint64, clusters []models.TransactionCluster) error
	// GetTransactionClusterByMember will return a transaction cluster that
	// contains the specified transaction ID as a member for the specified bank.
	// If no cluster can be found then nil and pg.NoRows will be returned
	// (wrapped).
	GetTransactionClusterByMember(ctx context.Context, bankAccountId uint64, transactionId uint64) (*models.TransactionCluster, error)

	// Teller functions
	CreateTellerLink(ctx context.Context, link *models.TellerLink) error
	UpdateTellerLink(ctx context.Context, link *models.TellerLink) error
	CreateTellerBankAccount(ctx context.Context, bankAccount *models.TellerBankAccount) error
	UpdateTellerBankAccount(ctx context.Context, bankAccount *models.TellerBankAccount) error
	CreateTellerTransaction(ctx context.Context, transaction *models.TellerTransaction) error
	CreateTellerSync(ctx context.Context, sync *models.TellerSync) error
	GetLatestTellerSync(ctx context.Context, tellerBankAccountId uint64) (*models.TellerSync, error)

	fileRepositoryInterface
}

type Repository interface {
	BaseRepository
	UserId() uint64

	GetMe(ctx context.Context) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
}

type UnauthenticatedRepository interface {
	CreateAccountV2(ctx context.Context, account *models.Account) error
	CreateLogin(ctx context.Context, email, password string, firstName, lastName string) (*models.Login, error)
	CreateUser(ctx context.Context, loginId, accountId uint64, user *models.User) error
	GetLinksForItem(ctx context.Context, itemId string) (*models.Link, error)
	GetLinkByTellerEnrollmentId(ctx context.Context, enrollmentId string) (*models.Link, error)
	GetLoginForEmail(ctx context.Context, emailAddress string) (*models.Login, error)
	ResetPassword(ctx context.Context, loginId uint64, hashedPassword string) error
	SetEmailVerified(ctx context.Context, emailAddress string) error
	UseBetaCode(ctx context.Context, betaId, usedBy uint64) error
	ValidateBetaCode(ctx context.Context, betaCode string) (*models.Beta, error)
}

func NewRepositoryFromSession(clock clock.Clock, userId, accountId uint64, database pg.DBI) Repository {
	return &repositoryBase{
		userId:    userId,
		accountId: accountId,
		txn:       database,
		clock:     clock,
	}
}

func NewUnauthenticatedRepository(clock clock.Clock, txn pg.DBI) UnauthenticatedRepository {
	return &unauthenticatedRepo{
		txn:   txn,
		clock: clock,
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

func (r *repositoryBase) GetIsSetup(ctx context.Context) (bool, error) {
	span := sentry.StartSpan(ctx, "GetIsSetup")
	defer span.Finish()

	return r.txn.ModelContext(span.Context(), &models.Link{}).
		Where(`"link"."account_id" = ?`, r.accountId).
		Exists()
}
