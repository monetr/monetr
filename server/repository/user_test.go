package repository_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	me, err := repo.GetMe(t.Context())
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
		err := unauthenticatedRepo.CreateAccountV2(t.Context(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		var ownerUser, memberUser models.User
		{ // Create the owner
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := unauthenticatedRepo.CreateLogin(t.Context(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")
			ownerUser = models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleOwner,
			}
			err = unauthenticatedRepo.CreateUser(t.Context(), &ownerUser)
			assert.NoError(t, err, "should successfully create user")
		}

		{ // Create the member
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := unauthenticatedRepo.CreateLogin(t.Context(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")
			memberUser = models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleMember,
			}
			err = unauthenticatedRepo.CreateUser(t.Context(), &memberUser)
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
			owner, err := ownerRepo.GetAccountOwner(t.Context())
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
			owner, err := memberRepo.GetAccountOwner(t.Context())
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
		err := unauthenticatedRepo.CreateAccountV2(t.Context(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		var memberUser models.User

		{ // Create the member
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := unauthenticatedRepo.CreateLogin(t.Context(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")
			memberUser = models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleMember,
			}
			err = unauthenticatedRepo.CreateUser(t.Context(), &memberUser)
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
			owner, err := memberRepo.GetAccountOwner(t.Context())
			assert.Error(t, err, "must return an error when there is no owner")
			assert.Nil(t, owner, "owner object should be nil when there is no owner")
		}
	})
}

func TestRepositoryBase_GetUserById(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
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

		result, err := repo.GetUserById(t.Context(), user.UserId)
		assert.NoError(t, err, "should not return an error for retrieving user by id")
		assert.Equal(t, user.UserId, result.UserId, "should be for the same user")
		assert.NotNil(t, result.Login, "login cannot be nil, it is used")
		assert.NotNil(t, result.Account, "account cannot be nil, it is used")
	})

	t.Run("user does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		// Create a repo for a user and account that do not exist in the database,
		// looking up any id in that session won't find a matching row.
		fakeUserId := models.NewID[models.User]()
		fakeAccountId := models.NewID[models.Account]()

		repo := repository.NewRepositoryFromSession(
			clock,
			fakeUserId,
			fakeAccountId,
			db,
			log,
		)

		result, err := repo.GetUserById(t.Context(), fakeUserId)
		assert.Error(t, err, "must return an error when the user does not exist")
		assert.Nil(t, result, "user should be nil when not found")
		assert.Contains(t, err.Error(), "user does not exist", "error should indicate the user was not found")
	})

	t.Run("is scoped to the account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		userOne, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		userTwo, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		// userOne belongs to a different account than userTwo's session, so the
		// lookup must come back empty even though the user id is real.
		repo := repository.NewRepositoryFromSession(
			clock,
			userTwo.UserId,
			userTwo.AccountId,
			db,
			log,
		)

		result, err := repo.GetUserById(t.Context(), userOne.UserId)
		assert.Error(t, err, "must not be able to read a user from a different account")
		assert.Nil(t, result, "user should be nil for a cross-account lookup")
	})

	t.Run("retrieves another user in the same account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		userOne, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		// Add a second user to the same account with their own login, this mirrors
		// what the fixtures do when they seed the first user.
		secondLogin, _ := fixtures.GivenIHaveLogin(t, clock)
		userTwo := models.User{
			LoginId:   secondLogin.LoginId,
			AccountId: userOne.AccountId,
			Role:      models.UserRoleMember,
		}
		unauthenticatedRepo := repository.NewUnauthenticatedRepository(clock, db)
		require.NoError(t, unauthenticatedRepo.CreateUser(t.Context(), &userTwo), "must be able to seed a second user")

		repo := repository.NewRepositoryFromSession(
			clock,
			userOne.UserId,
			userOne.AccountId,
			db,
			log,
		)

		// The lookup must honor the id argument, userOne needs to be able to read
		// userTwo's details for things like the Created By line in the UI.
		result, err := repo.GetUserById(t.Context(), userTwo.UserId)
		assert.NoError(t, err, "should be able to retrieve another user in the same account")
		assert.Equal(t, userTwo.UserId, result.UserId, "should return the requested user, not the session user")
		assert.Equal(t, secondLogin.LoginId, result.Login.LoginId, "should include the requested user's login")
	})

	t.Run("does not return the session user for a bogus id", func(t *testing.T) {
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

		// Even though the session user exists, asking for an id that doesn't must
		// not fall back to the session user.
		bogusId := models.NewID[models.User]()
		result, err := repo.GetUserById(t.Context(), bogusId)
		assert.Error(t, err, "must not fall back to the session user for a bogus id")
		assert.Nil(t, result, "user should be nil for a bogus id")
		assert.Contains(t, err.Error(), "user does not exist", "error should indicate the user was not found")
	})

	t.Run("includes login and account relations", func(t *testing.T) {
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

		result, err := repo.GetUserById(t.Context(), user.UserId)
		assert.NoError(t, err, "should not return an error")
		assert.NotNil(t, result.Login, "login relation should be eagerly loaded")
		assert.Equal(t, user.LoginId, result.Login.LoginId, "login should match the user's login")
		assert.NotNil(t, result.Account, "account relation should be eagerly loaded")
		assert.Equal(t, user.AccountId, result.Account.AccountId, "account should match the user's account")
	})
}
