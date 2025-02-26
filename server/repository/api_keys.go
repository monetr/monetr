package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/monetr/monetr/server/models"
)

type APIKeyRepository interface {
	CreateAPIKey(ctx context.Context, userId string, name string, expiresAt *time.Time) (string, *models.APIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error)
	ListAPIKeys(ctx context.Context, userId string) ([]models.APIKey, error)
	RevokeAPIKey(ctx context.Context, userId string, apiKeyId int64) error
	UpdateAPIKeyLastUsed(ctx context.Context, apiKeyId int64) error
}

func (r *repositoryBase) CreateAPIKey(ctx context.Context, userId string, name string, expiresAt *time.Time) (string, *models.APIKey, error) {
	// Generate a random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate API key: %w", err)
	}
	key := base64.URLEncoding.EncodeToString(keyBytes)
	
	// Hash the key for storage
	hash := sha256.Sum256([]byte(key))
	keyHash := base64.URLEncoding.EncodeToString(hash[:])

	apiKey := &models.APIKey{
		UserId:    userId,
		Name:      name,
		KeyHash:   keyHash,
		CreatedAt: time.Now().UTC(),
		IsActive:  true,
	}
	if expiresAt != nil {
		apiKey.ExpiresAt = *expiresAt
	}

	if _, err := r.txn.Model(apiKey).Insert(); err != nil {
		return "", nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return key, apiKey, nil
}

func (r *repositoryBase) GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	if err := r.txn.Model(&apiKey).Where("key_hash = ? AND is_active = TRUE", keyHash).Select(); err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	return &apiKey, nil
}

func (r *repositoryBase) ListAPIKeys(ctx context.Context, userId string) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	if err := r.txn.Model(&apiKeys).Where("user_id = ?", userId).Select(); err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	return apiKeys, nil
}

func (r *repositoryBase) RevokeAPIKey(ctx context.Context, userId string, apiKeyId int64) error {
	result, err := r.txn.Model(&models.APIKey{}).
		Where("user_id = ? AND api_key_id = ?", userId, apiKeyId).
		Set("is_active = FALSE").
		Update()
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("API key not found")
	}
	return nil
}

func (r *repositoryBase) UpdateAPIKeyLastUsed(ctx context.Context, apiKeyId int64) error {
	_, err := r.txn.Model(&models.APIKey{}).
		Where("api_key_id = ?", apiKeyId).
		Set("last_used_at = ?", time.Now().UTC()).
		Update()
	return err
}
