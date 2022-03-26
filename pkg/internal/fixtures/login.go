package fixtures

import (
	"context"
	"encoding/base32"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/hash"
	"github.com/monetr/monetr/pkg/internal/mock_stripe"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
	"github.com/xlzd/gotp"
)

func GivenIHaveLogin(t *testing.T) (_ models.Login, password string) {
	db := testutils.GetPgDatabase(t)

	repo := repository.NewUnauthenticatedRepository(db)

	password = gofakeit.Password(true, true, true, true, false, 64)
	email := testutils.GetUniqueEmail(t)
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()

	login, err := repo.CreateLogin(context.Background(), email, hash.HashPassword(email, password), firstName, lastName)
	require.NoError(t, err, "must be able to seed login")

	return *login, password
}

func GivenIHaveTOTPForLogin(t *testing.T, login *models.Login) *gotp.TOTP {
	db := testutils.GetPgDatabase(t)

	secret := base32.StdEncoding.EncodeToString([]byte(gofakeit.UUID()))
	loginTotp := gotp.NewDefaultTOTP(secret)
	login.TOTP = secret
	login.TOTPEnabledAt = myownsanity.TimeP(time.Now())
	result, err := db.Model(login).WherePK().Update(login)
	require.NoError(t, err, "must be able to update login with TOTP")
	require.Equal(t, 1, result.RowsAffected(), "must have only updated a single row")

	return loginTotp
}

func GivenIHaveABasicAccount(t *testing.T) (_ models.User, password string) {
	login, password := GivenIHaveLogin(t)
	user := GivenIHaveAnAccount(t, login)
	return user, password
}

func GivenIHaveAnAccount(t *testing.T, login models.Login) models.User {
	db := testutils.GetPgDatabase(t)
	repo := repository.NewUnauthenticatedRepository(db)
	subStatus := stripe.SubscriptionStatusActive
	account := models.Account{
		Timezone:                     gofakeit.TimeZoneRegion(),
		StripeCustomerId:             myownsanity.StringP(mock_stripe.FakeStripeCustomerId(t)),
		StripeSubscriptionId:         myownsanity.StringP(mock_stripe.FakeStripeSubscriptionId(t)),
		StripeWebhookLatestTimestamp: myownsanity.TimeP(time.Now().Add(-4 * time.Minute)),
		SubscriptionActiveUntil:      myownsanity.TimeP(time.Now().Add(10 * time.Minute)),
		SubscriptionStatus:           &subStatus,
	}
	err := repo.CreateAccountV2(context.Background(), &account)
	require.NoError(t, err, "must be able to seed basic account")

	user := models.User{
		LoginId:          login.LoginId,
		Login:            &login,
		AccountId:        account.AccountId,
		Account:          &account,
		FirstName:        login.FirstName,
		LastName:         login.LastName,
		StripeCustomerId: account.StripeCustomerId,
	}
	err = repo.CreateUser(context.Background(), login.LoginId, account.AccountId, &user)
	require.NoError(t, err, "must be able to see user for basic account")

	return user
}
