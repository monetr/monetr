package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/datasources/table"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validMapping(headers []string) table.Mapping {
	return table.Mapping{
		ID: table.IDSpec{
			Kind:   table.IDSpecKindNative,
			Fields: []table.FieldRef{{Name: "Id"}},
		},
		Amount: table.AmountSpec{
			Kind:   table.AmountKindSign,
			Fields: []table.FieldRef{{Name: "Amount"}},
		},
		Memo: table.FieldRef{Name: "Description"},
		Date: table.DateSpec{
			Fields: []table.FieldRef{{Name: "Date"}},
			Format: "YYYY-MM-DD",
		},
		Balance: table.BalanceSpec{
			Kind: table.BalanceKindNone,
		},
		Headers: headers,
	}
}

func TestRepositoryBase_CreateTransactionImportMapping(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repo := repository.NewRepositoryFromSession(
			mock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		headers := []string{"Date", "Description", "Amount", "Id"}
		mapping := models.TransactionImportMapping{
			Mapping: validMapping(headers),
		}

		err := repo.CreateTransactionImportMapping(context.Background(), &mapping)
		require.NoError(t, err, "must be able to create transaction import mapping")

		assert.False(t, mapping.TransactionImportMappingId.IsZero(), "id must be assigned")
		assert.Equal(t, user.AccountId, mapping.AccountId, "account id must be set from session")
		assert.Equal(t, user.UserId, mapping.CreatedBy, "created by must be set from session")
		assert.False(t, mapping.CreatedAt.IsZero(), "created at must be set")
		assert.False(t, mapping.UpdatedAt.IsZero(), "updated at must be set")
		assert.NotEmpty(t, mapping.Signature, "signature must be derived in BeforeInsert")

		stored := testutils.MustDBRead(t, mapping)
		assert.Equal(t, mapping.Signature, stored.Signature, "stored signature must match")
		assert.Equal(t, mapping.Mapping.Headers, stored.Mapping.Headers, "stored mapping headers must match")
	})
}

func TestRepositoryBase_GetTransactionImportMapping(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repo := repository.NewRepositoryFromSession(
			mock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		mapping := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &mapping))

		got, err := repo.GetTransactionImportMapping(context.Background(), mapping.TransactionImportMappingId)
		require.NoError(t, err, "must be able to retrieve mapping by id")
		assert.Equal(t, mapping.TransactionImportMappingId, got.TransactionImportMappingId)
		assert.Equal(t, mapping.Signature, got.Signature)
		assert.Equal(t, mapping.Mapping.Headers, got.Mapping.Headers)
	})

	t.Run("isolates by account", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		userA, _ := fixtures.GivenIHaveABasicAccount(t, mock)
		userB, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repoA := repository.NewRepositoryFromSession(
			mock, userA.UserId, userA.AccountId, testutils.GetPgDatabase(t), log,
		)
		repoB := repository.NewRepositoryFromSession(
			mock, userB.UserId, userB.AccountId, testutils.GetPgDatabase(t), log,
		)

		mapping := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
		}
		require.NoError(t, repoA.CreateTransactionImportMapping(context.Background(), &mapping))

		got, err := repoB.GetTransactionImportMapping(context.Background(), mapping.TransactionImportMappingId)
		require.Error(t, err, "must not return another account's mapping")
		assert.True(t, errors.Is(err, pg.ErrNoRows), "must surface a not-found error")
		assert.Nil(t, got, "must not return a record across accounts")
	})
}

func TestRepositoryBase_GetTransactionImportMappings(t *testing.T) {
	t.Run("lists all for account ordered by created_at desc", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repo := repository.NewRepositoryFromSession(
			mock, user.UserId, user.AccountId, testutils.GetPgDatabase(t), log,
		)

		first := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &first))

		mock.Add(1 * time.Minute)

		second := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Posted", "Memo", "Value", "Reference"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &second))

		got, err := repo.GetTransactionImportMappings(context.Background(), 10, 0)
		require.NoError(t, err, "must be able to list mappings")
		require.Len(t, got, 2, "must return both mappings")
		assert.Equal(t, second.TransactionImportMappingId, got[0].TransactionImportMappingId, "most recent must be first")
		assert.Equal(t, first.TransactionImportMappingId, got[1].TransactionImportMappingId)
	})

	t.Run("paginates results", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repo := repository.NewRepositoryFromSession(
			mock, user.UserId, user.AccountId, testutils.GetPgDatabase(t), log,
		)

		created := make([]models.ID[models.TransactionImportMapping], 0, 3)
		for i := 0; i < 3; i++ {
			mapping := models.TransactionImportMapping{
				Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
			}
			require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &mapping))
			created = append(created, mapping.TransactionImportMappingId)
			mock.Add(1 * time.Minute)
		}

		page1, err := repo.GetTransactionImportMappings(context.Background(), 2, 0)
		require.NoError(t, err)
		require.Len(t, page1, 2, "first page must return limit-many mappings")

		page2, err := repo.GetTransactionImportMappings(context.Background(), 2, 2)
		require.NoError(t, err)
		require.Len(t, page2, 1, "second page must return remaining mappings")

		assert.NotContains(t, []models.ID[models.TransactionImportMapping]{
			page1[0].TransactionImportMappingId,
			page1[1].TransactionImportMappingId,
		}, page2[0].TransactionImportMappingId, "pages must not overlap")
	})

	t.Run("isolates by account", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		userA, _ := fixtures.GivenIHaveABasicAccount(t, mock)
		userB, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repoA := repository.NewRepositoryFromSession(
			mock, userA.UserId, userA.AccountId, testutils.GetPgDatabase(t), log,
		)
		repoB := repository.NewRepositoryFromSession(
			mock, userB.UserId, userB.AccountId, testutils.GetPgDatabase(t), log,
		)

		mappingA := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
		}
		require.NoError(t, repoA.CreateTransactionImportMapping(context.Background(), &mappingA))

		mappingB := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Posted", "Memo", "Value", "Reference"}),
		}
		require.NoError(t, repoB.CreateTransactionImportMapping(context.Background(), &mappingB))

		got, err := repoA.GetTransactionImportMappings(context.Background(), 10, 0)
		require.NoError(t, err)
		require.Len(t, got, 1, "account A must only see its own mapping")
		assert.Equal(t, mappingA.TransactionImportMappingId, got[0].TransactionImportMappingId)
	})
}

func TestRepositoryBase_GetTransactionImportMappingsBySignature(t *testing.T) {
	t.Run("matches by signature", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repo := repository.NewRepositoryFromSession(
			mock, user.UserId, user.AccountId, testutils.GetPgDatabase(t), log,
		)

		matching1 := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &matching1))

		mock.Add(1 * time.Minute)

		matching2 := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &matching2))

		other := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Posted", "Memo", "Value", "Reference"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &other))

		assert.Equal(t, matching1.Signature, matching2.Signature, "same headers must produce same signature")
		assert.NotEqual(t, matching1.Signature, other.Signature, "different headers must produce different signatures")

		got, err := repo.GetTransactionImportMappingsBySignature(context.Background(), matching1.Signature, 10, 0)
		require.NoError(t, err)
		require.Len(t, got, 2, "must return only the two matching mappings")
		assert.Equal(t, matching2.TransactionImportMappingId, got[0].TransactionImportMappingId, "most recent must be first")
		assert.Equal(t, matching1.TransactionImportMappingId, got[1].TransactionImportMappingId)
	})

	t.Run("signature is case insensitive", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repo := repository.NewRepositoryFromSession(
			mock, user.UserId, user.AccountId, testutils.GetPgDatabase(t), log,
		)

		mixedCase := models.TransactionImportMapping{
			Mapping: validMapping([]string{"Date", "Description", "Amount", "Id"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &mixedCase))

		variant := models.TransactionImportMapping{
			Mapping: validMapping([]string{"DATE", "description", "Amount", "id"}),
		}
		require.NoError(t, repo.CreateTransactionImportMapping(context.Background(), &variant))

		assert.Equal(t, mixedCase.Signature, variant.Signature, "signature must ignore case differences in headers")
	})

	t.Run("isolates by account", func(t *testing.T) {
		mock := clock.NewMock()
		log := testutils.GetLog(t)
		userA, _ := fixtures.GivenIHaveABasicAccount(t, mock)
		userB, _ := fixtures.GivenIHaveABasicAccount(t, mock)

		repoA := repository.NewRepositoryFromSession(
			mock, userA.UserId, userA.AccountId, testutils.GetPgDatabase(t), log,
		)
		repoB := repository.NewRepositoryFromSession(
			mock, userB.UserId, userB.AccountId, testutils.GetPgDatabase(t), log,
		)

		headers := []string{"Date", "Description", "Amount", "Id"}
		mappingA := models.TransactionImportMapping{
			Mapping: validMapping(headers),
		}
		require.NoError(t, repoA.CreateTransactionImportMapping(context.Background(), &mappingA))

		mappingB := models.TransactionImportMapping{
			Mapping: validMapping(headers),
		}
		require.NoError(t, repoB.CreateTransactionImportMapping(context.Background(), &mappingB))

		require.Equal(t, mappingA.Signature, mappingB.Signature, "same headers across accounts must share signature")

		got, err := repoA.GetTransactionImportMappingsBySignature(context.Background(), mappingA.Signature, 10, 0)
		require.NoError(t, err)
		require.Len(t, got, 1, "account A must only see its own matching mapping")
		assert.Equal(t, mappingA.TransactionImportMappingId, got[0].TransactionImportMappingId)
	})
}
