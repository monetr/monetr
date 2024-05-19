package models

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/pkg/errors"
)

type Kind string

const (
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
	return string(i) == "" || strings.TrimPrefix(string(i), string(i.Kind())+"_") == ""
}

func (i ID[T]) Kind() Kind {
	inst := *new(T)
	prefix := inst.IdentityPrefix()
	return Kind(prefix)
}

var (
	entropy     io.Reader
	entropyOnce sync.Once
)

func cryptoEntropy() io.Reader {
	entropyOnce.Do(func() {
		entropy = &ulid.LockedMonotonicReader{
			MonotonicReader: ulid.Monotonic(rand.Reader, 0),
		}
	})
	return entropy
}

func NewID[T Identifiable](object *T) ID[T] {
	id := ulid.MustNew(ulid.Now(), cryptoEntropy())
	return ID[T](fmt.Sprintf(
		"%s_%s",
		(*object).IdentityPrefix(),
		strings.ToLower(id.String()),
	))
}

func ParseID[T Identifiable](input string) (ID[T], error) {
	inst := *new(T)
	prefix := inst.IdentityPrefix()

	if !strings.HasPrefix(input, prefix+"_") {
		return "", errors.Errorf("failed to parse ID for %T, expected prefix: %s ID: %s", inst, prefix, input)
	}

	return ID[T](input), nil
}
