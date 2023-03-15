package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/consts"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBaseGetLastPlaidSync(t *testing.T) {
	t.Run("no previous syncs", func(t *testing.T) {
		plaidLink := models.PlaidLink{
			ItemId:          gofakeit.UUID(),
			Products:        consts.PlaidProductStrings(),
			WebhookUrl:      "https://monetr.test/webhook",
			InstitutionId:   "ins_123",
			InstitutionName: "Platypus Bank",
		}
		testutils.MustDBInsert(t, &plaidLink)
		assert.NotZero(t, plaidLink.PlaidLinkID, "plaid link ID must not be zero, must have a valid record!")

		db := testutils.GetPgDatabase(t)
		repo := repository.NewRepositoryFromSession(123, 1234, db)
		result, err := repo.GetLastPlaidSync(context.Background(), plaidLink.PlaidLinkID)
		assert.NoError(t, err, "should not receive an error when there is no previous plaid sync")
		assert.Nil(t, result, "should receive nil, because there has not been a plaid sync before this")
	})

	t.Run("one previous sync", func(t *testing.T) {
		plaidLink := models.PlaidLink{
			ItemId:          gofakeit.UUID(),
			Products:        consts.PlaidProductStrings(),
			WebhookUrl:      "https://monetr.test/webhook",
			InstitutionId:   "ins_123",
			InstitutionName: "Platypus Bank",
		}
		testutils.MustDBInsert(t, &plaidLink)
		assert.NotZero(t, plaidLink.PlaidLinkID, "plaid link ID must not be zero, must have a valid record!")

		plaidSync := models.PlaidSync{
			PlaidLinkID: plaidLink.PlaidLinkID,
			Timestamp:   time.Now().UTC(),
			NextCursor:  gofakeit.UUID(),
			Trigger:     "webhook",
			Added:       12,
			Modified:    0,
			Removed:     0,
		}
		testutils.MustDBInsert(t, &plaidSync)
		assert.NotZero(t, plaidSync.PlaidSyncID, "plaid sync ID must not be zero, must have created a valid record!")

		db := testutils.GetPgDatabase(t)
		repo := repository.NewRepositoryFromSession(123, 1234, db)
		result, err := repo.GetLastPlaidSync(context.Background(), plaidLink.PlaidLinkID)
		assert.NoError(t, err, "should not receive an error when there is no previous plaid sync")
		assert.Equal(t, plaidSync.PlaidSyncID, result.PlaidSyncID, "should have received the last inserted plaid sync")
	})

	t.Run("multiple previous syncs", func(t *testing.T) {
		plaidLink := models.PlaidLink{
			ItemId:          gofakeit.UUID(),
			Products:        consts.PlaidProductStrings(),
			WebhookUrl:      "https://monetr.test/webhook",
			InstitutionId:   "ins_123",
			InstitutionName: "Platypus Bank",
		}
		testutils.MustDBInsert(t, &plaidLink)
		assert.NotZero(t, plaidLink.PlaidLinkID, "plaid link ID must not be zero, must have a valid record!")

		{ // One from a few days ago
			plaidSync := models.PlaidSync{
				PlaidLinkID: plaidLink.PlaidLinkID,
				Timestamp:   time.Now().AddDate(0, 0, -3).UTC(),
				NextCursor:  gofakeit.UUID(),
				Trigger:     "webhook",
				Added:       12,
				Modified:    0,
				Removed:     0,
			}
			testutils.MustDBInsert(t, &plaidSync)
			assert.NotZero(t, plaidSync.PlaidSyncID, "plaid sync ID must not be zero, must have created a valid record!")
		}

		// And one from yesterday
		plaidSync := models.PlaidSync{
			PlaidLinkID: plaidLink.PlaidLinkID,
			Timestamp:   time.Now().AddDate(0, 0, -1).UTC(),
			NextCursor:  gofakeit.UUID(),
			Trigger:     "webhook",
			Added:       4,
			Modified:    0,
			Removed:     0,
		}
		testutils.MustDBInsert(t, &plaidSync)
		assert.NotZero(t, plaidSync.PlaidSyncID, "plaid sync ID must not be zero, must have created a valid record!")

		db := testutils.GetPgDatabase(t)
		repo := repository.NewRepositoryFromSession(123, 1234, db)
		result, err := repo.GetLastPlaidSync(context.Background(), plaidLink.PlaidLinkID)
		assert.NoError(t, err, "should not receive an error when there is no previous plaid sync")
		assert.Equal(t, plaidSync.PlaidSyncID, result.PlaidSyncID, "should have received the last inserted plaid sync")
	})
}
