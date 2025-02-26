package models

import (
	"time"
)

type APIKey struct {
	tableName struct{} `pg:"api_keys"`

	APIKeyId   int64     `json:"apiKeyId" pg:"api_key_id,pk"`
	UserId     string    `json:"-" pg:"user_id"`
	Name       string    `json:"name" pg:"name"`
	KeyHash    string    `json:"-" pg:"key_hash"`
	CreatedAt  time.Time `json:"createdAt" pg:"created_at"`
	LastUsedAt time.Time `json:"lastUsedAt,omitempty" pg:"last_used_at"`
	ExpiresAt  time.Time `json:"expiresAt,omitempty" pg:"expires_at"`
	IsActive   bool      `json:"isActive" pg:"is_active"`
}
