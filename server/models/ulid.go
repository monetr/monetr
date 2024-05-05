package models

import (
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
)

type Kind string

const (
	UnknownIDKind            Kind = ""
	LoginIDKind              Kind = "lgn"
	AccountIDKind            Kind = "acct"
	UserIDKind               Kind = "user"
	LinkIDKind               Kind = "link"
	BetaIDKind               Kind = "beta"
	SecretIDKind             Kind = "scrt"
	BankAccountIDKind        Kind = "bac"
	SpendingIDKind           Kind = "spnd"
	FundingScheduleIDKind    Kind = "fund"
	FileIDKind               Kind = "file"
	JobIDKind                Kind = "job"
	TransactionIDKind        Kind = "txn"
	TransactionClusterIDKind Kind = "tcl"
	PlaidLinkIDKind          Kind = "plx"
	PlaidSyncIDKind          Kind = "psyn"
	PlaidBankAccountIDKind   Kind = "pbac"
	PlaidTransactionIDKind   Kind = "ptxn"
)

type Identifiable interface {
	IdentityPrefix() string
}

type Identifier interface {
	Kind() Kind
	String() string
}

type ID[T Identifiable] string

func (i ID[T]) String() string {
	return string(i)
}

func (i ID[T]) IsZero() bool {
	return string(i) == ""
}

func (i ID[T]) Kind() Kind {
	str := i.String()
	index := strings.LastIndex(str, "_")
	if index == -1 {
		return UnknownIDKind
	}

	return Kind(str[:index])
}

func NewID[T Identifiable](object *T) ID[T] {
	id := ulid.Make()
	return ID[T](fmt.Sprintf(
		"%s_%s",
		(*object).IdentityPrefix(),
		strings.ToLower(id.String()),
	))
}
