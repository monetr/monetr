package fixtures

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestGivenIHaveLogin(t *testing.T) {
	clock := clock.NewMock()
	login, password := GivenIHaveLogin(t, clock)
	assert.NotEmpty(t, password, "password cannot be empty")
	assert.NotZero(t, login.LoginId, "login must have been created")
}

func TestGivenIHaveABasicAccount(t *testing.T) {
	clock := clock.NewMock()
	user, password := GivenIHaveABasicAccount(t, clock)
	assert.NotEmpty(t, password, "password cannot be empty")
	assert.NotZero(t, user.UserId, "user Id must be present")
	assert.NotNil(t, user.Account, "account must be present")
	assert.NotZero(t, user.AccountId, "account Id must be present")
	assert.NotNil(t, user.Login, "login must be present")
	assert.NotZero(t, user.LoginId, "login Id must be present")

	location, err := time.LoadLocation(user.Account.Timezone)
	assert.NoError(t, err, "account must have valid location")
	assert.NotNil(t, location, "location cannot be nil")

	assert.True(t, user.Account.IsSubscriptionActive(clock.Now()), "account subscription must be active")
}

func TestGivenIHaveTOTPForLogin(t *testing.T) {
	clock := clock.NewMock()
	login, _ := GivenIHaveLogin(t, clock)
	assert.NotZero(t, login.LoginId, "login must have been created")
	assert.Empty(t, login.TOTP, "should not have a TOTP initially")
	assert.Nil(t, login.TOTPEnabledAt, "TOTP enabled at should be nil")

	loginTotp := GivenIHaveTOTPForLogin(t, clock, &login)
	assert.NotNil(t, loginTotp, "should return a TOTP object")
	assert.NotEmpty(t, login.TOTP, "should now have a TOTP secret for the login")
	assert.NotNil(t, login.TOTPEnabledAt, "TOTP enabled at should no longer be nil")
}
