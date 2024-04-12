package identifier

import (
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
)

type Kind string

const (
	LoginKind              Kind = "login"
	AccountKind            Kind = "account"
	UserKind               Kind = "user"
	LinkKind               Kind = "link"
	PlaidLinkKind          Kind = "plaid_link"
	TellerLinkKind         Kind = "teller_link"
	BankAccountKind        Kind = "bank_account"
	PlaidBankAccountKind   Kind = "plaid_bank_account"
	TellerBankAccountKind  Kind = "teller_bank_account"
	FundingSchedulekind    Kind = "funding_schedule"
	TransactionKind        Kind = "transaction"
	PlaidTransactionKind   Kind = "plaid_transaction"
	TellerTransactionKind  Kind = "teller_transaction"
	TransactionClusterKind Kind = "transaction_cluster"
	SecretKind             Kind = "secret"
	SpendingKind           Kind = "spending"
)

type Identifier interface {
	Kind() Kind
	String() string
}

type ID string

func (i ID) String() string {
	return string(i)
}

func (i ID) Kind() Kind {
	str := i.String()
	index := strings.LastIndex(str, "_")
	return Kind(str[:index])
}

func New(kind Kind) ID {
	id := ulid.Make()
	return ID(fmt.Sprintf("%s_%s", kind, strings.ToLower(id.String())))
}
