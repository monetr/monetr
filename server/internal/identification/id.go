package identification

import (
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
)

type Kind string

const (
	LoginKind              Kind = "login"
	UserKind               Kind = "user"
	AccountKind            Kind = "account"
	LinkKind               Kind = "link"
	PlaidLinkKind          Kind = "plaid_link"
	TellerLinkKind         Kind = "teller_link"
	BankAccountKind        Kind = "bank_account"
	PlaidBankAccountKind   Kind = "plaid_bank_account"
	TellerBankAccountKind  Kind = "teller_bank_account"
	PlaidSyncKind          Kind = "plaid_sync"
	TellerSyncKind         Kind = "teller_sync"
	TransactionKind        Kind = "transaction"
	PlaidTransactionKind   Kind = "plaid_transaction"
	TellerTransactionKind  Kind = "teller_transaction"
	TransactionClusterKind Kind = "transaction_cluster"
	SecretKind             Kind = "secret"
	SpendingKind           Kind = "spending"
	FundingScheduleKind    Kind = "funding_schedule"
	FileKind               Kind = "file"
	CronJobKind            Kind = "cron"
	JobKind                Kind = "job"
	BetaKind               Kind = "beta"
)

type ID string

func (i ID) Kind() Kind {
	str := string(i)
	index := strings.LastIndex(str, "_")
	return Kind(str[:index])
}

func New(kind Kind) ID {
	return ID(strings.ToLower(fmt.Sprintf("%s_%s", kind, ulid.Make())))
}
