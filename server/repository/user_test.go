package repository

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_GetMe(t *testing.T) {
	clock := clock.NewMock()
	db := testutils.GetPgDatabase(t)

	user, _ := testutils.SeedAccount(t, db, clock, testutils.WithPlaidAccount)

	repo := NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

	me, err := repo.GetMe(context.Background())
	assert.NoError(t, err, "should not return an error for retrieving me")
	assert.Equal(t, user.UserId, me.UserId, "should be for the same user")
	assert.NotNil(t, me.Login, "login cannot be nil, it is used")
	assert.NotNil(t, me.Account, "account cannot be nil, it is used")
}
