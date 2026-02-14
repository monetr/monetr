package models_test

import (
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestFile_GetStorePath(t *testing.T) {
	t.Run("will generate a good path", func(t *testing.T) {
		file := models.File{
			FileId:      models.ID[models.File]("file_01kheqbr884scw3wt8v70g95yk"),
			AccountId:   models.ID[models.Account]("acct_01kheqbr884scw3wt8vacxg1sd"),
			Kind:        "transactions/uploads",
			Name:        "transactions (10).ofx",
			ContentType: models.IntuitQFXContentType,
			Size:        100,
		}
		path, err := file.GetStorePath()
		assert.NoError(t, err, "must not return an error generating a path for a valid file")
		assert.Equal(t, "data/transactions/uploads/61/01kheqbr884scw3wt8vacxg1sd/66/01kheqbr884scw3wt8v70g95yk", path)
	})

	t.Run("will fail if the file ID is not valid", func(t *testing.T) {
		file := models.File{
			FileId:      models.ID[models.File](""),
			AccountId:   models.ID[models.Account]("acct_01kheqbr884scw3wt8vacxg1sd"),
			Kind:        "transactions/uploads",
			Name:        "transactions (10).ofx",
			ContentType: models.IntuitQFXContentType,
			Size:        100,
		}
		path, err := file.GetStorePath()
		assert.EqualError(t, err, "no valid file ID")
		assert.Empty(t, path, "path should be empty if there is an error")
	})

	t.Run("will fail if the account ID is not valid", func(t *testing.T) {
		file := models.File{
			FileId:      models.ID[models.File]("file_01kheqbr884scw3wt8v70g95yk"),
			AccountId:   models.ID[models.Account](""),
			Kind:        "transactions/uploads",
			Name:        "transactions (10).ofx",
			ContentType: models.IntuitQFXContentType,
			Size:        100,
		}
		path, err := file.GetStorePath()
		assert.EqualError(t, err, "no valid account ID")
		assert.Empty(t, path, "path should be empty if there is an error")
	})
}
