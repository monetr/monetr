package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
)

// NewAPIKeyRepository creates a new repository for managing API keys.
func NewAPIKeyRepository(db pg.DBI) APIKeyRepository {
	return &apiKeyRepository{
		db: db,
	}
}

type apiKeyRepository struct {
	db pg.DBI
}

func (r *apiKeyRepository) CreateAPIKey(ctx context.Context, userId string, name string, expiresAt *time.Time) (string, *models.APIKey, error) {
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

	if _, err := r.db.ModelContext(ctx, apiKey).Insert(); err != nil {
		return "", nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return key, apiKey, nil
}

func (r *apiKeyRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	apiKey := &models.APIKey{
		KeyHash: keyHash,
	}

	err := r.db.ModelContext(ctx, apiKey).
		Where("key_hash = ?", keyHash).
		Where("is_active = ?", true).
		Select()
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (r *apiKeyRepository) ListAPIKeys(ctx context.Context, userId string) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	err := r.db.ModelContext(ctx, &apiKeys).
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Select()
	if err != nil {
		return nil, err
	}
	
	return apiKeys, nil
}

func (r *apiKeyRepository) RevokeAPIKey(ctx context.Context, userId string, apiKeyId int64) error {
	result, err := r.db.ModelContext(ctx, &models.APIKey{}).
		Set("is_active = ?", false).
		Where("user_id = ?", userId).
		Where("api_key_id = ?", apiKeyId).
		Update()
	if err != nil {
		return err
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no API key found with ID %d for user %s", apiKeyId, userId)
	}
	
	return nil
}

func (r *apiKeyRepository) UpdateAPIKeyLastUsed(ctx context.Context, apiKeyId int64) error {
	_, err := r.db.ModelContext(ctx, &models.APIKey{}).
		Set("last_used_at = ?", time.Now()).
		Where("api_key_id = ?", apiKeyId).
		Update()
	return err
}
