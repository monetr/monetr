package models

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// Encrypted stores one encrypted value on a model, you must embed it on the
// model directly and not use it as a struct field.
type Encrypted[T any] struct {
	KeyEncryptionKeyId ID[KeyEncryptionKey] `json:"-" bun:"key_encryption_key_id,notnull"`
	KeyEncryptionKey   *KeyEncryptionKey    `json:"-" bun:"rel:belongs-to,join:key_encryption_key_id=key_encryption_key_id"`
	// EncryptedData is hex encoded: format version, wrapped DEK, nonce, and
	// ciphertext.
	EncryptedData string `json:"-" bun:"encrypted_data,notnull"`
}

// EncryptionAAD builds the additional authenticated data for the encrypted
// value stored on model. It's the bun table name followed by every primary key
// column and its value, sorted by column name:
//
//	plaid_links:account_id=acct_01J...:plaid_link_id=plx_01J...
func EncryptionAAD[T any](db bun.IDB, model *T) ([]byte, error) {
	value := reflect.ValueOf(model)
	if value.Elem().Kind() != reflect.Struct {
		return nil, errors.Errorf("encryption AAD requires a pointer to a struct, got %T", model)
	}
	value = value.Elem()

	table := db.Dialect().Tables().Get(value.Type())
	if len(table.PKs) == 0 {
		return nil, errors.Errorf("table %s has no primary key to build an encryption AAD from", table.Name)
	}

	pks := slices.Clone(table.PKs)
	slices.SortFunc(pks, func(a, b *schema.Field) int {
		return strings.Compare(a.Name, b.Name)
	})

	var aad strings.Builder
	aad.WriteString(table.Name)
	for _, pk := range pks {
		if pk.HasZeroValue(value) {
			return nil, errors.Errorf(
				"primary key %s.%s must be set before encrypting",
				table.Name, pk.Name,
			)
		}

		part := fmt.Sprint(pk.Value(value).Interface())
		// ':' separates parts and '=' separates column from value, so neither may
		// appear in a value or two different rows could produce the same AAD.
		// monetr IDs never contain either.
		if strings.ContainsAny(part, ":=") {
			return nil, errors.Errorf(
				"primary key %s.%s contains a reserved character",
				table.Name, pk.Name,
			)
		}

		aad.WriteString(":")
		aad.WriteString(pk.Name)
		aad.WriteString("=")
		aad.WriteString(part)
	}

	return []byte(aad.String()), nil
}
