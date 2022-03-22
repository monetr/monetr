package fixtures

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGivenIHaveLogin(t *testing.T) {
	login, password := GivenIHaveLogin(t)
	assert.NotEmpty(t, password, "password cannot be empty")
	assert.NotZero(t, login.LoginId, "login must have been created")
}

func TestGivenIHaveABasicAccount(t *testing.T) {
	user, password := GivenIHaveABasicAccount(t)
	assert.NotEmpty(t, password, "password cannot be empty")
	assert.NotZero(t, user.UserId, "user Id must be present")
	assert.NotNil(t, user.Account, "account must be present")
	assert.NotZero(t, user.AccountId, "account Id must be present")
	assert.NotNil(t, user.Login, "login must be present")
	assert.NotZero(t, user.LoginId, "login Id must be present")

	location, err := time.LoadLocation(user.Account.Timezone)
	assert.NoError(t, err, "account must have valid location")
	assert.NotNil(t, location, "location cannot be nil")

	assert.True(t, user.Account.IsSubscriptionActive(), "account subscription must be active")
}

func TestGivenIHaveTOTPForLogin(t *testing.T) {
	login, _ := GivenIHaveLogin(t)
	assert.NotZero(t, login.LoginId, "login must have been created")
	assert.Empty(t, login.TOTP, "should not have a TOTP initially")
	assert.Nil(t, login.TOTPEnabledAt, "TOTP enabled at should be nil")

	loginTotp := GivenIHaveTOTPForLogin(t, &login)
	assert.NotNil(t, loginTotp, "should return a TOTP object")
	assert.NotEmpty(t, login.TOTP, "should now have a TOTP secret for the login")
	assert.NotNil(t, login.TOTPEnabledAt, "TOTP enabled at should no longer be nil")
}
