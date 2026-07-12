package models

type EncryptedData struct {
	EncryptedDataVersion            uint64               `json:"-" pg:"encrypted_data_version,notnull"`
	EncryptedDataKeyEncryptionKeyId ID[KeyEncryptionKey] `json:"-" pg:"key_encryption_key_id,notnull"`
	EncryptedDataKeyEncryptionKey   *KeyEncryptionKey    `json:"-" pg:"rel:has-one"`
	EncryptedDataWrapNonce          []byte               `json:"-" pg:"encrypted_data_wrapped_nonce,notnull"`
	EncryptedDataWrappedDek         []byte               `json:"-" pg:"encrypted_data_wrapped_dek,notnull"`
	EncryptedDataNonce              []byte               `json:"-" pg:"encrypted_data_nonce,notnull"`
	EncryptedDataCiphertext         []byte               `json:"-" pg:"encrypted_data_ciphertext,notnull"`
}
