package models

type KeyEncryptionKey struct {
	KeyEncryptionKeyId ID[KeyEncryptionKey] `json:"keyEncryptionKeyId"`
}

func (KeyEncryptionKey) IdentityPrefix() string {
	return "kek"
}
