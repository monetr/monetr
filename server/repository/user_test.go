package repository_test

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_GetMe(t *testing.T) {
	clock := clock.NewMock()
	db := testutils.GetPgDatabase(t)

	user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

	repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

	me, err := repo.GetMe(context.Background())
	assert.NoError(t, err, "should not return an error for retrieving me")
	assert.Equal(t, user.UserId, me.UserId, "should be for the same user")
	assert.NotNil(t, me.Login, "login cannot be nil, it is used")
	assert.NotNil(t, me.Account, "account cannot be nil, it is used")
}
