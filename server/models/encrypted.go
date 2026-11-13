package models

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
	"golang.org/x/crypto/chacha20poly1305"
)

const encryptedFormatV1 byte = 0x01

// DataKeyWrapper wraps and unwraps DEKs with a KEK. The keyring implements it.
// It is defined here so models doesn't import the encryption package, and so
// raw KEK bytes never leave the keyring.
type DataKeyWrapper interface {
	// WrapDataKey wraps dek with the active KEK and returns that KEK's ID along
	// with the wrapped DEK.
	WrapDataKey(
		ctx context.Context,
		dek []byte,
	) (ID[KeyEncryptionKey], []byte, error)
	// UnwrapDataKey unwraps a DEK that was wrapped by the given KEK.
	UnwrapDataKey(
		ctx context.Context,
		kekId ID[KeyEncryptionKey],
		wrapped []byte,
	) ([]byte, error)
}

// Encrypted stores one encrypted value on a model, you must embed it on the
// model directly and not use it as a struct field.
type Encrypted[T any] struct {
	KeyEncryptionKeyId ID[KeyEncryptionKey] `json:"-" bun:"key_encryption_key_id,notnull"`
	KeyEncryptionKey   *KeyEncryptionKey    `json:"-" bun:"rel:belongs-to,join:key_encryption_key_id=key_encryption_key_id"`
	// EncryptedData is hex encoded: format version, wrapped DEK, nonce, and
	// ciphertext.
	EncryptedData string `json:"-" bun:"encrypted_data,notnull"`
}

func (e *Encrypted[T]) Encrypt(
	ctx context.Context,
	keys DataKeyWrapper,
	aad []byte,
	value T,
) error {
	plaintext, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to serialize value for encryption")
	}
	defer clear(plaintext)

	dek := make([]byte, chacha20poly1305.KeySize)
	rand.Read(dek)
	defer clear(dek)

	aead, err := chacha20poly1305.NewX(dek)
	if err != nil {
		return errors.Wrap(err, "failed to create data cipher")
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	rand.Read(nonce)
	ciphertext := aead.Seal(nil, nonce, plaintext, aad)

	kekId, wrappedDEK, err := keys.WrapDataKey(ctx, dek)
	if err != nil {
		return errors.Wrap(err, "failed to wrap data key")
	}

	e.KeyEncryptionKeyId = kekId
	e.EncryptedData = encryptedBlob{
		wrappedDEK: wrappedDEK,
		nonce:      nonce,
		ciphertext: ciphertext,
	}.encode()
	return nil
}

func (e *Encrypted[T]) Decrypt(
	ctx context.Context,
	keys DataKeyWrapper,
	aad []byte,
) (T, error) {
	return *new(T), nil
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

// Blob layout: [1B format version][4B wrapped DEK length, big endian uint32]
// [wrapped DEK][24B nonce][ciphertext || tag]. The length prefix means this
// code doesn't need to know how the keyring formats wrapped DEKs.
type encryptedBlob struct {
	wrappedDEK []byte
	nonce      []byte
	ciphertext []byte
}

func (b encryptedBlob) encode() string {
	buf := make([]byte, 0, 5+len(b.wrappedDEK)+len(b.nonce)+len(b.ciphertext))
	buf = append(buf, encryptedFormatV1)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(b.wrappedDEK)))
	buf = append(buf, b.wrappedDEK...)
	buf = append(buf, b.nonce...)
	buf = append(buf, b.ciphertext...)
	return hex.EncodeToString(buf)
}
