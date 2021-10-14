package tasks

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/internal/mock_stripe"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepairPlaidWebhooksTask_Execute(t *testing.T) {
	type Seed struct {
		User       models.User
		Account    models.Account
		Links      []models.Link
		PlaidLinks []models.PlaidLink
	}

	seedData := func(t *testing.T, numberOfActiveAccounts, numberOfInactiveAccounts int) ([]Seed, *pg.DB) {
		db := testutils.GetPgDatabase(t)

		login := models.Login{
			Email:           testutils.GivenIHaveAnEmail(t),
			FirstName:       gofakeit.FirstName(),
			LastName:        gofakeit.LastName(),
			IsEnabled:       true,
			IsEmailVerified: true,
		}
		{ // Create the login.
			result, err := db.Model(&login).Insert(&login)
			require.NoError(t, err, "must be able to seed login for test")
			assert.Equal(t, 1, result.RowsAffected(), "should only have 1 row affected")
		}

		seeds := make([]Seed, 0, numberOfInactiveAccounts+numberOfActiveAccounts)
		for a := 0; a < numberOfActiveAccounts; a++ {
			seed := Seed{
				Account: models.Account{
					Timezone:                     time.Local.String(),
					StripeCustomerId:             myownsanity.StringP(mock_stripe.FakeStripeCustomerId(t)),
					StripeSubscriptionId:         myownsanity.StringP(mock_stripe.FakeStripeSubscriptionId(t)),
					StripeWebhookLatestTimestamp: myownsanity.TimeP(time.Now().Add(-10 * time.Minute)),
					SubscriptionActiveUntil:      myownsanity.TimeP(time.Now().Add(1 * time.Hour)),
				},
				Links: []models.Link{
					{
						LinkType:              models.ManualLinkType,
						LinkStatus:            models.LinkStatusSetup,
						InstitutionName:       "I Am A Bank",
						CustomInstitutionName: "But not a real bank",
						LastSuccessfulUpdate:  nil,
						BankAccounts:          nil,
					},
				},
				PlaidLinks: nil,
			}
			db.
		}
	}

	t.Run("all accounts", func(t *testing.T) {
		options :=

	})
}
