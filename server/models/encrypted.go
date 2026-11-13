package models

type Encrypted[T any] struct {
	KeyEncryptionKeyId ID[KeyEncryptionKey] `json:"-" bun:"TODO FILL IN FK HERE!"`
	// EncryptedData contains the wrapped DEK, a nonce and the ciphertext for the
	// data being stored.
	EncryptedData string `json:"-" bun:"encrypted_data,notnull"`
}
