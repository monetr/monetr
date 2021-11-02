package repository

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaidRepositoryBase_GetLink(t *testing.T) {
	repo := GetTestAuthenticatedRepository(t)
	txn := repo.(*repositoryBase).txn
	require.NotNil(t, txn, "must be able to pull the transaction for test")

	plaidLink := &models.PlaidLink{
		ItemId: gofakeit.UUID(),
		Products: []string{
			"transactions",
		},
		WebhookUrl: "https://monetr.test/webhook",
	}

	link := &models.Link{
		AccountId:             repo.AccountId(),
		LinkType:              models.PlaidLinkType,
		PlaidLinkId:           nil,
		LinkStatus:            models.LinkStatusSetup,
		InstitutionName:       "Institution " + t.Name(),
		CustomInstitutionName: "Institution " + t.Name(),
		CreatedAt:             time.Now(),
		CreatedByUserId:       repo.UserId(),
		UpdatedAt:             time.Now(),
		LastSuccessfulUpdate:  nil,
	}

	{ // Create the links.
		require.NoError(t, repo.CreatePlaidLink(context.Background(), plaidLink), "must create plaid link")
		link.PlaidLinkId = &plaidLink.PlaidLinkID
		require.NoError(t, repo.CreateLink(context.Background(), link), "must create link")
	}

	plaidRepo := NewPlaidRepository(txn)

	t.Run("simple", func(t *testing.T) {
		readLink, err := plaidRepo.GetLink(context.Background(), link.AccountId, link.LinkId)
		assert.NoError(t, err, "failed to retrieve link")
		assert.NotNil(t, readLink.PlaidLink, "must include plaid link child")
		assert.EqualValues(t, link.LinkId, readLink.LinkId, "link Id must match")
		assert.EqualValues(t, plaidLink.PlaidLinkID, readLink.PlaidLink.PlaidLinkID, "plaid link Id must match")
	})

	t.Run("not found", func(t *testing.T) {
		readLink, err := plaidRepo.GetLink(context.Background(), link.AccountId, link.LinkId + 100)
		assert.EqualError(t, err, "failed to retrieve link: pg: no rows in result set")
		assert.Nil(t, readLink, "link must be nil")
	})
}

func TestPlaidRepositoryBase_GetLinkByItemId(t *testing.T) {
	repo := GetTestAuthenticatedRepository(t)
	txn := repo.(*repositoryBase).txn
	require.NotNil(t, txn, "must be able to pull the transaction for test")

	plaidLink := &models.PlaidLink{
		ItemId: gofakeit.UUID(),
		Products: []string{
			"transactions",
		},
		WebhookUrl: "https://monetr.test/webhook",
	}

	link := &models.Link{
		AccountId:             repo.AccountId(),
		LinkType:              models.PlaidLinkType,
		PlaidLinkId:           nil,
		LinkStatus:            models.LinkStatusSetup,
		InstitutionName:       "Institution " + t.Name(),
		CustomInstitutionName: "Institution " + t.Name(),
		CreatedAt:             time.Now(),
		CreatedByUserId:       repo.UserId(),
		UpdatedAt:             time.Now(),
		LastSuccessfulUpdate:  nil,
	}

	{ // Create the links.
		require.NoError(t, repo.CreatePlaidLink(context.Background(), plaidLink), "must create plaid link")
		link.PlaidLinkId = &plaidLink.PlaidLinkID
		require.NoError(t, repo.CreateLink(context.Background(), link), "must create link")
	}

	plaidRepo := NewPlaidRepository(txn)

	t.Run("simple", func(t *testing.T) {
		readLink, err := plaidRepo.GetLinkByItemId(context.Background(), plaidLink.ItemId)
		assert.NoError(t, err, "failed to retrieve link")
		assert.NotNil(t, readLink.PlaidLink, "must include plaid link child")
		assert.EqualValues(t, link.LinkId, readLink.LinkId, "link Id must match")
		assert.EqualValues(t, plaidLink.PlaidLinkID, readLink.PlaidLink.PlaidLinkID, "plaid link Id must match")
	})

	t.Run("not found", func(t *testing.T) {
		readLink, err := plaidRepo.GetLinkByItemId(context.Background(), "not a real item id")
		assert.EqualError(t, err, "failed to retrieve link by item Id: pg: no rows in result set")
		assert.Nil(t, readLink, "link must be nil")
	})
}
