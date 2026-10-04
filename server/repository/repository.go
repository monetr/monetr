package repository

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/uptrace/bun"
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

	// CreatePlaidTransactions takes a Plaid transaction model and ensures the
	// account ID and the created at timestamp are properly set then stores the
	// transaction in the database.
	CreatePlaidTransactions(ctx context.Context, transactions ...*PlaidTransaction) error

	DeleteFundingSchedule(ctx context.Context, bankAccountId ID[BankAccount], fundingScheduleId ID[FundingSchedule]) error
	DeletePlaidLink(ctx context.Context, plaidLinkId ID[PlaidLink]) error
	DeleteSpending(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending]) error
	DeleteTransaction(ctx context.Context, bankAccountId ID[BankAccount], transactionId ID[Transaction]) error
	SoftDeleteTransaction(ctx context.Context, bankAccountId ID[BankAccount], transactionId ID[Transaction]) error
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
	// GetTransactionsForSimilarity will return all of the non-deleted
	// transactions for an account in date ascending order. This is the opposite
	// of what [BaseRepository.GetTransactions] does, and this function also does
	// not take any pagination parameters and returns all of the data in a single
	// chunk. This is intended only for use within similarity calculations.
	GetTransactionsForSimilarity(
		ctx context.Context,
		bankAccountId ID[BankAccount],
	) ([]Transaction, error)
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
	// GetTransactionsByLunchFlowId is meant to be used by the lunch flow sync
	// code and will retrieve all of the transactions that already exist in the
	// database for a given sync.
	GetTransactionsByLunchFlowId(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		lunchFlowIds []string,
	) (map[string]Transaction, error)

	// Deprecated: Use GetTransactionsByPlaidId
	GetTransactionsByPlaidTransactionId(ctx context.Context, linkId ID[Link], plaidTransactionIds []string) ([]Transaction, error)
	GetTransactionsForSpending(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending], limit, offset int) ([]Transaction, error)
	GetTransactionsForRecurring(ctx context.Context, bankAccountId ID[BankAccount], transactionRecurringId ID[TransactionRecurring], limit, offset int) ([]Transaction, error)
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

	// GetClusteredTransactions returns every transaction in the bank account
	// that currently belongs to a cluster, including soft deleted ones. Only the
	// primary key and the cluster ID are populated.
	GetClusteredTransactions(
		ctx context.Context,
		bankAccountId ID[BankAccount],
	) ([]Transaction, error)
	// UpsertTransactionClusters will insert or update the provided clusters by
	// their primary key. It doesn't delete anything, use DeleteTransactionClusters
	// for that. The name on an existing cluster is never overwritten since that's
	// for the user to set.
	UpsertTransactionClusters(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		clusters []TransactionCluster,
	) error
	// DeleteTransactionClusters removes the specified clusters by their IDs.
	// Transactions referencing these clusters will have their cluster ID set to
	// null.
	DeleteTransactionClusters(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		clusterIds []ID[TransactionCluster],
	) error
	// UpdateTransactionClusterIds sets the cluster ID on each of the provided
	// transactions by their primary key. A nil cluster ID removes the
	// transaction from whatever cluster it was in. No other columns are written.
	UpdateTransactionClusterIds(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactions []Transaction,
	) error
	// GetTransactionClusterByMember will return a transaction cluster that
	// contains the specified transaction ID as a member for the specified bank.
	// If no cluster can be found then nil and sql.ErrNoRows will be returned
	// (wrapped).
	// Deprecated: You should simply read the cluster ID on the transaction instead.
	GetTransactionClusterByMember(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionId ID[Transaction],
	) (*TransactionCluster, error)
	GetTransactionCluster(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionClusterId ID[TransactionCluster],
	) (*TransactionCluster, error)
	// GetTransactionClusterIds returns the ID of every transaction cluster for
	// the specified bank account, ordered by ID.
	GetTransactionClusterIds(
		ctx context.Context,
		bankAccountId ID[BankAccount],
	) ([]ID[TransactionCluster], error)
	GetTransactionsByCluster(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionClusterId ID[TransactionCluster],
		limit, offset int,
	) ([]Transaction, error)

	GetTransactionRecurringById(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionRecurringId ID[TransactionRecurring],
	) (*TransactionRecurring, error)

	// GetTransactionRecurringByCluster returns the recurring transactions for the
	// specified cluster, there is at most one for each direction.
	GetTransactionRecurringByCluster(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionClusterId ID[TransactionCluster],
	) ([]TransactionRecurring, error)
	// UpsertTransactionRecurring will insert or update the provided recurring
	// transactions by their cluster and direction. An existing recurring
	// transaction for the same cluster and direction is updated in place and
	// keeps its ID. It doesn't delete anything, use DeleteTransactionRecurring
	// for that.
	UpsertTransactionRecurring(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		recurring []TransactionRecurring,
	) error
	// UpdateTransactionRecurringIds sets the recurring ID on each of the provided
	// transactions by their primary key. A nil recurring ID removes the
	// transaction from whatever recurring transaction it was part of. No other
	// columns are written.
	UpdateTransactionRecurringIds(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactions []Transaction,
	) error
	// DeleteTransactionRecurring removes the specified recurring transactions by
	// their IDs. Transactions referencing them will have their recurring ID set
	// to null.
	DeleteTransactionRecurring(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionRecurringIds []ID[TransactionRecurring],
	) error

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

	CreateTransactionImportMapping(
		ctx context.Context,
		mapping *TransactionImportMapping,
	) error
	GetTransactionImportMapping(
		ctx context.Context,
		mappingId ID[TransactionImportMapping],
	) (*TransactionImportMapping, error)
	GetTransactionImportMappings(
		ctx context.Context,
		limit, offset int,
	) ([]TransactionImportMapping, error)
	GetTransactionImportMappingsBySignature(
		ctx context.Context,
		signature string,
		limit, offset int,
	) ([]TransactionImportMapping, error)

	GetTransactionImport(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionImportId ID[TransactionImport],
	) (*TransactionImport, error)
	CreateTransactionImport(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionImport *TransactionImport,
	) error
	UpdateTransactionImport(
		ctx context.Context,
		bankAccountId ID[BankAccount],
		transactionImport *TransactionImport,
	) error

	CreateLunchFlowLink(ctx context.Context, link *LunchFlowLink) error
	UpdateLunchFlowLink(ctx context.Context, link *LunchFlowLink) error
	RemoveLunchFlowLink(ctx context.Context, id ID[LunchFlowLink]) error
	GetLunchFlowLinks(ctx context.Context) ([]LunchFlowLink, error)
	GetLunchFlowLink(ctx context.Context, id ID[LunchFlowLink]) (*LunchFlowLink, error)

	CreateLunchFlowBankAccount(ctx context.Context, bankAccount *LunchFlowBankAccount) error
	GetLunchFlowBankAccountsByLunchFlowLink(
		ctx context.Context, id ID[LunchFlowLink],
	) ([]LunchFlowBankAccount, error)
	GetLunchFlowBankAccount(
		ctx context.Context,
		id ID[LunchFlowBankAccount],
	) (*LunchFlowBankAccount, error)
	GetLunchFlowBankAccountForLunchFlowLink(
		ctx context.Context,
		linkId ID[LunchFlowLink],
		id ID[LunchFlowBankAccount],
	) (*LunchFlowBankAccount, error)
	UpdateLunchFlowBankAccount(
		ctx context.Context,
		bankAccount *LunchFlowBankAccount,
	) error

	CreateLunchFlowTransactions(ctx context.Context, transaction []LunchFlowTransaction) error

	fileRepositoryInterface

	GetApiKeyById(ctx context.Context, id ID[ApiKey]) (*ApiKey, error)
	GetApiKeys(ctx context.Context) ([]ApiKey, error)
	CreateApiKey(ctx context.Context, key *ApiKey) error
	DeleteApiKey(ctx context.Context, id ID[ApiKey]) error

	GetUserById(
		ctx context.Context,
		id ID[User],
	) (*User, error)
}

type Repository interface {
	BaseRepository
	UserId() ID[User]

	GetMe(ctx context.Context) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
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
	GetApiKey(ctx context.Context, keyId ID[ApiKey]) (*ApiKey, error)
}

func NewRepositoryFromSession(
	clock clock.Clock,
	userId ID[User],
	accountId ID[Account],
	database bun.IDB,
	log *slog.Logger,
) Repository {
	return &repositoryBase{
		userId:    userId,
		accountId: accountId,
		txn:       database,
		clock:     clock,
		log:       log,
	}
}

func NewUnauthenticatedRepository(
	clock clock.Clock,
	txn bun.IDB,
) UnauthenticatedRepository {
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

	return r.txn.NewSelect().Model(&Link{}).
		Where(`"link"."account_id" = ?`, r.accountId).
		Where(`"link"."deleted_at" IS NULL`).
		Exists(span.Context())
}
