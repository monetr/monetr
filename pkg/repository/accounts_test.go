package repository

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func TestRepositoryBase_GetAccount(t *testing.T) {
	testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
		repo := GetTestAuthenticatedRepository(t, db)
		account, err := repo.GetAccount(context.Background())
		assert.NoError(t, err, "must be able to retrieve the current account")
		assert.NotNil(t, account, "account object must not be nil")
	})
}
