package repository

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
)

type BaseRepository interface {
	AccountId() ID[Account]

	CreatePlaidBankAccount(ctx context.Context, bankAccount *PlaidBankAccount) error
	UpdatePlaidBankAccount(ctx context.Context, bankAccount *PlaidBankAccount) error

	GetPlaidBankAccountsByLinkId(ctx context.Context, linkId ID[Link]) ([]PlaidBankAccount, error)

	AddExpenseToTransaction(ctx context.Context, transaction *Transaction, spending *Spending) error
	CreateBankAccounts(ctx context.Context, bankAccounts ...*BankAccount) error
	CreateFundingSchedule(ctx context.Context, fundingSchedule *FundingSchedule) error
	CreateLink(ctx context.Context, link *Link) error
	CreatePlaidLink(ctx context.Context, link *PlaidLink) error
	CreateSpending(ctx context.Context, expense *Spending) error
	CreateTransaction(ctx context.Context, bankAccountId ID[BankAccount], transaction *Transaction) error

	// CreatePlaidTransaction takes a Plaid transaction model and ensures the
	// account ID and the created at timestamp are properly set then stores the
	// transaction in the database.
	CreatePlaidTransaction(ctx context.Context, transaction *PlaidTransaction) error
	CreatePlaidTransactions(ctx context.Context, transactions ...*PlaidTransaction) error

	DeleteFundingSchedule(ctx context.Context, bankAccountId ID[BankAccount], fundingScheduleId ID[FundingSchedule]) error
	DeletePlaidLink(ctx context.Context, plaidLinkId ID[PlaidLink]) error
	DeleteSpending(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending]) error
	DeleteTransaction(ctx context.Context, bankAccountId ID[BankAccount], transactionId ID[Transaction]) error
	GetAccount(ctx context.Context) (*Account, error)
	// GetAccountOwner will return a User object for the currently authenticated
	// account, as well as the Login and Account sub object for that user. If one
	// is not found then an error is returned.
	GetAccountOwner(ctx context.Context) (*User, error)
	GetBalances(ctx context.Context, bankAccountId ID[BankAccount]) (*Balances, error)
	GetBankAccount(ctx context.Context, bankAccountId ID[BankAccount]) (*BankAccount, error)
	GetBankAccounts(ctx context.Context) ([]BankAccount, error)
	GetBankAccountsByLinkId(ctx context.Context, linkId ID[Link]) ([]BankAccount, error)
	// GetBankAccountsWithPlaidByLinkId will return all the bank accounts
	// associated with the provided link ID that also have a Plaid bank account
	// associated with them.
	GetBankAccountsWithPlaidByLinkId(ctx context.Context, linkId ID[Link]) ([]BankAccount, error)
	GetFundingSchedule(ctx context.Context, bankAccountId ID[BankAccount], fundingScheduleId ID[FundingSchedule]) (*FundingSchedule, error)
	GetFundingSchedules(ctx context.Context, bankAccountId ID[BankAccount]) ([]FundingSchedule, error)
	GetIsSetup(ctx context.Context) (bool, error)
	GetLink(ctx context.Context, linkId ID[Link]) (*Link, error)
	GetLinkIsManual(ctx context.Context, linkId ID[Link]) (bool, error)
	GetLinkIsManualByBankAccountId(ctx context.Context, bankAccountId ID[BankAccount]) (bool, error)
	GetLinks(ctx context.Context) ([]Link, error)

	// Plaid syncing
	GetLastPlaidSync(ctx context.Context, plaidLinkId ID[PlaidLink]) (*PlaidSync, error)
	RecordPlaidSync(ctx context.Context, plaidLinkId ID[PlaidLink], trigger, nextCursor string, added, modified, removed int) error

	GetNumberOfPlaidLinks(ctx context.Context) (int, error)
	GetSpending(ctx context.Context, bankAccountId ID[BankAccount]) ([]Spending, error)
	GetSpendingByFundingSchedule(ctx context.Context, bankAccountId ID[BankAccount], fundingScheduleId ID[FundingSchedule]) ([]Spending, error)
	GetSpendingById(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending]) (*Spending, error)
	GetSpendingExists(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending]) (bool, error)
	GetTransaction(ctx context.Context, bankAccountId ID[BankAccount], transactionId ID[Transaction]) (*Transaction, error)
	GetTransactions(ctx context.Context, bankAccountId ID[BankAccount], limit, offset int) ([]Transaction, error)
	// GetTransactionsAfter will return all of the transactions after the
	// specified date, if the specified date is null then all transactions for an
	// account is returned. This is intended to be used for partial syncing for
	// file uploads or teller.
	GetTransactionsAfter(ctx context.Context, bankAccountId ID[BankAccount], after *time.Time) ([]Transaction, error)
	// GetPendingTransactions is the same as GetTransactions but will only return
	// transactions that are currently in a pending state. It will not return
	// transactions that have been deleted.
	GetPendingTransactions(ctx context.Context, bankAccountId ID[BankAccount], limit, offset int) ([]Transaction, error)
	// GetRecentDepositTransactions will return all deposit transactions for the specified bank account within the past
	// 24 hours.
	GetRecentDepositTransactions(ctx context.Context, bankAccountId ID[BankAccount]) ([]Transaction, error)
	GetTransactionsByPlaidId(ctx context.Context, linkId ID[Link], plaidTransactionIds []string) (map[string]Transaction, error)

	// GetTransactonsByUploadIdentifier is meant to be used by the file import
	// processing code. It will retrieve transactions that already exist in the
	// database by their external upload identifier.
	GetTransactonsByUploadIdentifier(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		uploadIdentifiers []string,
	) (map[string]Transaction, error)

	// Deprecated: Use GetTransactionsByPlaidId
	GetTransactionsByPlaidTransactionId(ctx context.Context, linkId ID[Link], plaidTransactionIds []string) ([]Transaction, error)
	GetTransactionsForSpending(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending], limit, offset int) ([]Transaction, error)
	InsertTransactions(ctx context.Context, transactions []Transaction) error
	ProcessTransactionSpentFrom(ctx context.Context, bankAccountId ID[BankAccount], input, existing *Transaction) (updatedExpenses []Spending, _ error)
	UpdateBankAccount(ctx context.Context, bankAccount *BankAccount) error
	UpdateSpending(ctx context.Context, bankAccountId ID[BankAccount], updates []Spending) error
	UpdateLink(ctx context.Context, link *Link) error
	UpdateFundingSchedule(ctx context.Context, fundingSchedule *FundingSchedule) error
	UpdatePlaidLink(ctx context.Context, plaidLink *PlaidLink) error
	UpdateTransaction(ctx context.Context, bankAccountId ID[BankAccount], transaction *Transaction) error

	// UpdateTransactions is unique in that it REQUIRES that all data on each transaction object be populated. It is
	// doing a bulk update, so if data is missing it has the potential to overwrite a transaction incorrectly.
	UpdateTransactions(ctx context.Context, transactions []*Transaction) error

	// WriteTransactionClusters will take the array of transaction clusters
	// provided and persist them to the trnasaction clusters table but will also
	// delete any existing transaction clusters for the bank account specified.
	// This is because clusters are meant to be regenerated each time new
	// transactions come in.
	WriteTransactionClusters(ctx context.Context, bankAccountId ID[BankAccount], clusters []TransactionCluster) error
	// GetTransactionClusterByMember will return a transaction cluster that
	// contains the specified transaction ID as a member for the specified bank.
	// If no cluster can be found then nil and pg.NoRows will be returned
	// (wrapped).
	GetTransactionClusterByMember(ctx context.Context, bankAccountId ID[BankAccount], transactionId ID[Transaction]) (*TransactionCluster, error)

	GetTransactionUpload(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionUploadId ID[TransactionUpload],
	) (*TransactionUpload, error)
	CreateTransactionUpload(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionUpload *TransactionUpload,
	) error

	fileRepositoryInterface
}

type Repository interface {
	BaseRepository
	UserId() ID[User]
	GetMe(ctx context.Context) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	
	// API Key methods
	CreateAPIKey(ctx context.Context, userId string, name string, expiresAt *time.Time) (string, *APIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error)
	ListAPIKeys(ctx context.Context, userId string) ([]APIKey, error)
	RevokeAPIKey(ctx context.Context, userId string, apiKeyId int64) error
	UpdateAPIKeyLastUsed(ctx context.Context, apiKeyId int64) error
}

type UnauthenticatedRepository interface {
	CreateAccountV2(ctx context.Context, account *Account) error
	CreateLogin(ctx context.Context, email, password string, firstName, lastName string) (*Login, error)
	CreateUser(ctx context.Context, user *User) error
	GetLinksForItem(ctx context.Context, itemId string) (*Link, error)
	GetLoginForEmail(ctx context.Context, emailAddress string) (*Login, error)
	ResetPassword(ctx context.Context, loginId ID[Login], hashedPassword string) error
	SetEmailVerified(ctx context.Context, emailAddress string) error
	UseBetaCode(ctx context.Context, betaId ID[Beta], usedBy ID[User]) error
	ValidateBetaCode(ctx context.Context, betaCode string) (*Beta, error)
}

func NewRepositoryFromSession(clock clock.Clock, userId ID[User], accountId ID[Account], database pg.DBI) Repository {
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

func (r *repositoryBase) UserId() ID[User] {
	return r.userId
}

func (r *repositoryBase) AccountId() ID[Account] {
	return r.accountId
}

func (r *repositoryBase) AccountIdStr() string {
	return r.AccountId().String()
}

func (r *repositoryBase) GetIsSetup(ctx context.Context) (bool, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	return r.txn.ModelContext(span.Context(), &Link{}).
		Where(`"link"."account_id" = ?`, r.accountId).
		Where(`"link"."deleted_at" IS NULL`).
		Exists()
}
