package myownsanity_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestLeftJoin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		type User struct {
			ID        int
			AccountID int
		}
		type Account struct {
			ID int
		}
		accounts := []Account{
			{
				ID: 1,
			},
			{
				ID: 2,
			},
			{
				ID: 3,
			},
		}
		users := []User{
			{
				ID:        1,
				AccountID: 1,
			},
			{
				ID:        2,
				AccountID: 1,
			},
			{
				ID:        3,
				AccountID: 2,
			},
		}

		result := myownsanity.LeftJoin(accounts, users, func(a Account, b User) bool {
			return a.ID == b.AccountID
		})
		assert.Equal(t, 1, result[0].From.ID)
		assert.Equal(t, 1, result[0].Join[0].ID)
		assert.Equal(t, 1, result[0].Join[0].AccountID)
		assert.Equal(t, 2, result[0].Join[1].ID)
		assert.Equal(t, 1, result[0].Join[1].AccountID)
		assert.Equal(t, 2, result[1].From.ID)
		assert.Equal(t, 3, result[1].Join[0].ID)
		assert.Equal(t, 2, result[1].Join[0].AccountID)
		assert.Equal(t, 2, result[1].Join[0].AccountID)
		assert.Equal(t, 3, result[2].From.ID)
		assert.Empty(t, result[2].Join)
	})
}
