package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_GetMe(t *testing.T) {
	clock := clock.NewMock()
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)

	user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

	repo := repository.NewRepositoryFromSession(
		clock,
		user.UserId,
		user.AccountId,
		db,
		log,
	)

	me, err := repo.GetMe(context.Background())
	assert.NoError(t, err, "should not return an error for retrieving me")
	assert.Equal(t, user.UserId, me.UserId, "should be for the same user")
	assert.NotNil(t, me.Login, "login cannot be nil, it is used")
	assert.NotNil(t, me.Account, "account cannot be nil, it is used")
}

func TestRepositoryBase_GetAccountOwner(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		unauthenticatedRepo := GetTestUnauthenticatedRepository(t, clock)

		account := models.Account{
			Timezone:                time.UTC.String(),
			Locale:                  "en_US",
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err := unauthenticatedRepo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		var ownerUser, memberUser models.User
		{ // Create the owner
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := unauthenticatedRepo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")
			ownerUser = models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleOwner,
			}
			err = unauthenticatedRepo.CreateUser(context.Background(), &ownerUser)
			assert.NoError(t, err, "should successfully create user")
		}

		{ // Create the member
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := unauthenticatedRepo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")
			memberUser = models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleMember,
			}
			err = unauthenticatedRepo.CreateUser(context.Background(), &memberUser)
			assert.NoError(t, err, "should successfully create user")
		}

		{ // When we are authenticated as the owner
			ownerRepo := repository.NewRepositoryFromSession(
				clock,
				ownerUser.UserId,
				ownerUser.AccountId,
				db,
				log,
			)
			owner, err := ownerRepo.GetAccountOwner(context.Background())
			assert.NoError(t, err, "must be able to retrieve the owner")
			assert.NotNil(t, owner.Account, "account sub object should be included")
			assert.NotNil(t, owner.Login, "account sub object should be included")
			assert.Equal(t, ownerUser.UserId, owner.UserId, "should match the owner we created")
		}

		{ // When we are authenticated as a member
			memberRepo := repository.NewRepositoryFromSession(
				clock,
				memberUser.UserId,
				memberUser.AccountId,
				db,
				log,
			)
			// Even if we are the member, we should still retrieve the owner who is
			// not us. Makes sure the current user ID doesn't change the query.
			owner, err := memberRepo.GetAccountOwner(context.Background())
			assert.NoError(t, err, "must be able to retrieve the owner")
			assert.NotNil(t, owner.Account, "account sub object should be included")
			assert.NotNil(t, owner.Login, "account sub object should be included")
			assert.Equal(t, ownerUser.UserId, owner.UserId, "should match the owner we created")
		}
	})

	t.Run("missing owner", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		unauthenticatedRepo := GetTestUnauthenticatedRepository(t, clock)

		account := models.Account{
			Timezone:                time.UTC.String(),
			Locale:                  "en_US",
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err := unauthenticatedRepo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		var memberUser models.User

		{ // Create the member
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := unauthenticatedRepo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")
			memberUser = models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleMember,
			}
			err = unauthenticatedRepo.CreateUser(context.Background(), &memberUser)
			assert.NoError(t, err, "should successfully create user")
		}

		{ // When there is no owner, we should get an error
			memberRepo := repository.NewRepositoryFromSession(
				clock,
				memberUser.UserId,
				memberUser.AccountId,
				db,
				log,
			)
			owner, err := memberRepo.GetAccountOwner(context.Background())
			assert.Error(t, err, "must return an error when there is no owner")
			assert.Nil(t, owner, "owner object should be nil when there is no owner")
		}
	})
}
