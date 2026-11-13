package models

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
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
	var result T
	blob, err := parseEncryptedBlob(e.EncryptedData)
	if err != nil {
		return result, err
	}

	dek, err := keys.UnwrapDataKey(ctx, e.KeyEncryptionKeyId, blob.wrappedDEK)
	if err != nil {
		return result, errors.Wrap(err, "failed to unwrap data key")
	}
	defer clear(dek)

	aead, err := chacha20poly1305.NewX(dek)
	if err != nil {
		return result, errors.Wrap(err, "failed to create data cipher")
	}

	plaintext, err := aead.Open(nil, blob.nonce, blob.ciphertext, aad)
	if err != nil {
		return result, errors.Wrap(err, "failed to decrypt value")
	}
	defer clear(plaintext)

	if err := json.Unmarshal(plaintext, &result); err != nil {
		return result, errors.Wrap(err, "failed to deserialize decrypted value")
	}
	return result, nil
}

// EncryptionAAD builds the additional authenticated data for the encrypted
// value stored on model. It's a URL with the bun table name as the host and
// every primary key column and its value as query parameters. url.Values
// escapes the values and sorts the parameters by column name:
//
//	aad://plaid_links?account_id=acct_01J...&plaid_link_id=plx_01J...
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

	query := url.Values{}
	for _, pk := range table.PKs {
		if pk.HasZeroValue(value) {
			return nil, errors.Errorf(
				"primary key %s.%s must be set before encrypting",
				table.Name, pk.Name,
			)
		}

		query.Set(pk.Name, fmt.Sprint(pk.Value(value).Interface()))
	}

	aad := url.URL{
		Scheme:   "aad",
		Host:     table.Name,
		RawQuery: query.Encode(),
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

func parseEncryptedBlob(data string) (encryptedBlob, error) {
	raw, err := hex.DecodeString(data)
	if err != nil {
		return encryptedBlob{}, errors.Wrap(err, "encrypted data is not valid hex")
	}
	if len(raw) < 5 {
		return encryptedBlob{}, errors.New("encrypted data is too short")
	}
	if raw[0] != encryptedFormatV1 {
		return encryptedBlob{}, errors.Errorf("unsupported encrypted data format: %d", raw[0])
	}

	length := binary.BigEndian.Uint32(raw[1:5])
	rest := raw[5:]
	// Compare as uint64 before converting so a corrupt length can't overflow int
	// on 32-bit platforms.
	minimum := uint64(length) +
		chacha20poly1305.NonceSizeX +
		chacha20poly1305.Overhead
	if uint64(len(rest)) < minimum {
		return encryptedBlob{}, errors.New("encrypted data is truncated")
	}
	dekLength := int(length)

	return encryptedBlob{
		wrappedDEK: rest[:dekLength],
		nonce:      rest[dekLength : dekLength+chacha20poly1305.NonceSizeX],
		ciphertext: rest[dekLength+chacha20poly1305.NonceSizeX:],
	}, nil
}
