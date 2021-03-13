package repository

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

type Repository interface {
	AccountId() uint64
	UserId() uint64

	CreateBankAccounts(bankAccounts ...models.BankAccount) error
	CreateExpense(expense *models.Expense) error
	CreateFundingSchedule(fundingSchedule *models.FundingSchedule) error
	CreateLink(link *models.Link) error
	CreatePlaidLink(link *models.PlaidLink) error
	GetAccount() (*models.Account, error)
	GetBankAccounts() ([]models.BankAccount, error)
	GetBankAccountsByLinkId(linkId uint64) ([]models.BankAccount, error)
	GetExpenses(bankAccountId uint64) ([]models.Expense, error)
	GetFundingSchedules(bankAccountId uint64) ([]models.FundingSchedule, error)
	GetFundingSchedule(bankAccountId, fundingScheduleId uint64) (*models.FundingSchedule, error)
	GetIsSetup() (bool, error)
	GetJob(jobId string) (models.Job, error)
	GetLink(linkId uint64) (*models.Link, error)
	GetLinkIsManual(linkId uint64) (bool, error)
	GetLinkIsManualByBankAccountId(bankAccountId uint64) (bool, error)
	GetLinks() ([]models.Link, error)
	GetMe() (*models.User, error)
	GetTransactions(bankAccountId uint64, limit, offset int) ([]models.Transaction, error)
	GetTransactionsByPlaidId(linkId uint64, plaidTransactionIds []string) (map[string]TransactionUpdateId, error)
	InsertTransactions(transactions []models.Transaction) error
	UpdateBankAccounts(accounts []models.BankAccount) error
	UpdateLink(link *models.Link) error
}

type UnauthenticatedRepository interface {
	CreateLogin(email, hashedPassword string, isEnabled bool) (*models.Login, error)
	CreateAccount(timezone *time.Location) (*models.Account, error)
	CreateUser(loginId, accountId uint64, firstName, lastName string) (*models.User, error)
	CreateRegistration(loginId uint64) (*models.Registration, error)

	// VerifyRegistration takes a registrationId and will finalize the registration record. If the registration has
	// already been completed an error is returned.
	VerifyRegistration(registrationId string) (*models.User, error)
}

func NewRepositoryFromSession(userId, accountId uint64, txn *pg.Tx) Repository {
	return &repositoryBase{
		userId:    userId,
		accountId: accountId,
		txn:       txn,
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

func (r *repositoryBase) GetMe() (*models.User, error) {
	var user models.User
	err := r.txn.Model(&user).
		Relation("Login").
		Relation("Login.EmailVerifications").
		Relation("Login.PhoneVerifications").
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

