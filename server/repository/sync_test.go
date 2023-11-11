package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBaseGetLastPlaidSync(t *testing.T) {
	t.Run("no previous syncs", func(t *testing.T) {
		clock := clock.NewMock()
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
		repo := repository.NewRepositoryFromSession(clock, 123, 1234, db)
		result, err := repo.GetLastPlaidSync(context.Background(), plaidLink.PlaidLinkID)
		assert.NoError(t, err, "should not receive an error when there is no previous plaid sync")
		assert.Nil(t, result, "should receive nil, because there has not been a plaid sync before this")
	})

	t.Run("one previous sync", func(t *testing.T) {
		clock := clock.NewMock()
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
			Timestamp:   clock.Now().UTC(),
			NextCursor:  gofakeit.UUID(),
			Trigger:     "webhook",
			Added:       12,
			Modified:    0,
			Removed:     0,
		}
		testutils.MustDBInsert(t, &plaidSync)
		assert.NotZero(t, plaidSync.PlaidSyncID, "plaid sync ID must not be zero, must have created a valid record!")

		db := testutils.GetPgDatabase(t)
		repo := repository.NewRepositoryFromSession(clock, 123, 1234, db)
		result, err := repo.GetLastPlaidSync(context.Background(), plaidLink.PlaidLinkID)
		assert.NoError(t, err, "should not receive an error when there is no previous plaid sync")
		assert.Equal(t, plaidSync.PlaidSyncID, result.PlaidSyncID, "should have received the last inserted plaid sync")
	})

	t.Run("multiple previous syncs", func(t *testing.T) {
		clock := clock.NewMock()
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
				Timestamp:   clock.Now().UTC(),
				NextCursor:  gofakeit.UUID(),
				Trigger:     "webhook",
				Added:       12,
				Modified:    0,
				Removed:     0,
			}
			testutils.MustDBInsert(t, &plaidSync)
			assert.NotZero(t, plaidSync.PlaidSyncID, "plaid sync ID must not be zero, must have created a valid record!")
		}

		// Move the clock forward 2 days
		clock.Add(2 * 24 * time.Hour)

		// And one from yesterday
		plaidSync := models.PlaidSync{
			PlaidLinkID: plaidLink.PlaidLinkID,
			Timestamp:   clock.Now().UTC(),
			NextCursor:  gofakeit.UUID(),
			Trigger:     "webhook",
			Added:       4,
			Modified:    0,
			Removed:     0,
		}
		testutils.MustDBInsert(t, &plaidSync)
		assert.NotZero(t, plaidSync.PlaidSyncID, "plaid sync ID must not be zero, must have created a valid record!")

		clock.Add(1 * 24 * time.Hour)

		db := testutils.GetPgDatabase(t)
		repo := repository.NewRepositoryFromSession(clock, 123, 1234, db)
		result, err := repo.GetLastPlaidSync(context.Background(), plaidLink.PlaidLinkID)
		assert.NoError(t, err, "should not receive an error when there is no previous plaid sync")
		assert.Equal(t, plaidSync.PlaidSyncID, result.PlaidSyncID, "should have received the last inserted plaid sync")
	})
}
