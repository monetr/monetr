package fixtures

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/mock_stripe"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v81"
	"github.com/xlzd/gotp"
)

func GivenIHaveLogin(t *testing.T, clock clock.Clock) (_ models.Login, password string) {
	db := testutils.GetPgDatabase(t)

	repo := repository.NewUnauthenticatedRepository(clock, db)

	password = gofakeit.Password(true, true, true, true, false, 64)
	email := testutils.GetUniqueEmail(t)
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()

	login, err := repo.CreateLogin(context.Background(), email, password, firstName, lastName)
	require.NoError(t, err, "must be able to seed login")

	return *login, password
}

func GivenIHaveTOTPForLogin(t *testing.T, clock clock.Clock, login *models.Login) *gotp.TOTP {
	db := testutils.GetPgDatabase(t)

	secureRepo := repository.NewSecurityRepository(db, clock)
	uri, recoveryCodes, err := secureRepo.SetupTOTP(context.Background(), login.LoginId)
	require.NoError(t, err, "Must be able to setup TOTP on login")
	require.NotEmpty(t, uri, "Must return a TOTP URI")
	require.NotEmpty(t, recoveryCodes, "Must return an array of TOTP recovery codes")

	parsed, err := url.Parse(uri)
	require.NoError(t, err, "Must be able to parse the TOTP URI")
	require.NotNil(t, parsed)
	require.Equal(t, "otpauth", parsed.Scheme)
	require.Equal(t, "totp", parsed.Host)
	require.Equal(t, "monetr", parsed.Query().Get("issuer"))

	secret := parsed.Query().Get("secret")
	require.NotEmpty(t, secret, "must have a secret in the URI")

	loginTotp := gotp.NewDefaultTOTP(secret)

	err = secureRepo.EnableTOTP(
		context.Background(),
		login.LoginId,
		loginTotp.AtTime(clock.Now()),
	)
	require.NoError(t, err, "Must be able to enable TOTP for login")

	*login = testutils.MustRetrieve(t, *login)

	return loginTotp
}

func GivenIHaveTOTPCodeForLogin(t *testing.T, clock clock.Clock, login *models.Login) string {
	loginTotp := GivenIHaveTOTPForLogin(t, clock, login)
	code := loginTotp.Now()
	// If the code would change very soon, then use the next code instead.
	futureTimestamp := clock.Now().Add(1 * time.Second).Unix()
	if loginTotp.At(futureTimestamp) != code {
		code = loginTotp.At(futureTimestamp)
	}

	return code
}

func GivenIHaveABasicAccount(t *testing.T, clock clock.Clock) (_ models.User, password string) {
	login, password := GivenIHaveLogin(t, clock)
	user := GivenIHaveAnAccount(t, clock, login)
	return user, password
}

func GivenIHaveAnAccount(t *testing.T, clock clock.Clock, login models.Login) models.User {
	db := testutils.GetPgDatabase(t)
	repo := repository.NewUnauthenticatedRepository(clock, db)
	subStatus := stripe.SubscriptionStatusActive
	account := models.Account{
		Timezone:                     gofakeit.TimeZoneRegion(),
		Locale:                       "en_US",
		StripeCustomerId:             myownsanity.Pointer(mock_stripe.FakeStripeCustomerId(t)),
		StripeSubscriptionId:         myownsanity.Pointer(mock_stripe.FakeStripeSubscriptionId(t)),
		StripeWebhookLatestTimestamp: myownsanity.Pointer(clock.Now().Add(-4 * time.Minute)),
		SubscriptionActiveUntil:      myownsanity.Pointer(clock.Now().Add(10 * time.Minute)),
		SubscriptionStatus:           &subStatus,
		TrialEndsAt:                  nil,
	}
	err := repo.CreateAccountV2(context.Background(), &account)
	require.NoError(t, err, "must be able to seed basic account")

	user := models.User{
		LoginId:   login.LoginId,
		Login:     &login,
		AccountId: account.AccountId,
		Account:   &account,
		Role:      models.UserRoleOwner,
	}
	err = repo.CreateUser(context.Background(), &user)
	require.NoError(t, err, "must be able to see user for basic account")

	return user
}

func GivenIHaveATrialingAccount(t *testing.T, clock clock.Clock, login models.Login) models.User {
	db := testutils.GetPgDatabase(t)
	repo := repository.NewUnauthenticatedRepository(clock, db)
	account := models.Account{
		Timezone:                     gofakeit.TimeZoneRegion(),
		Locale:                       "en_US",
		StripeCustomerId:             nil,
		StripeSubscriptionId:         nil,
		StripeWebhookLatestTimestamp: nil,
		SubscriptionActiveUntil:      nil,
		SubscriptionStatus:           nil,
		TrialEndsAt:                  myownsanity.TimeP(clock.Now().AddDate(0, 0, 30)),
	}
	err := repo.CreateAccountV2(context.Background(), &account)
	require.NoError(t, err, "must be able to seed basic account")

	user := models.User{
		LoginId:   login.LoginId,
		Login:     &login,
		AccountId: account.AccountId,
		Account:   &account,
		Role:      models.UserRoleOwner,
	}
	err = repo.CreateUser(context.Background(), &user)
	require.NoError(t, err, "must be able to see user for basic account")

	return user
}

func GivenAccountIsInTimezone(t *testing.T, account *models.Account, location *time.Location) {
	db := testutils.GetPgDatabase(t)
	result, err := db.Model(account).
		WherePK().
		Set(`"timezone" = ?`, location.String()).
		UpdateNotZero(account)
	require.NoError(t, err, "must be able to set timezone")
	require.EqualValues(t, 1, result.RowsAffected(), "must have updated a single row")
}
