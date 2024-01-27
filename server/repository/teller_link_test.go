package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepository_CreateTellerLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

		link := models.TellerLinkWithAccessToken{
			TellerLink: models.TellerLink{
				EnrollmentId:         gofakeit.UUID(),
				UserId:               gofakeit.Generate("user_######"),
				Status:               models.TellerLinkStatusSetup,
				ErrorCode:            nil,
				InstitituionName:     fmt.Sprintf("Bank Of %s", gofakeit.City()),
				LastManualSync:       nil,
				LastSuccessfulUpdate: nil,
				LastAttemptedUpdate:  nil,
			},
			AccessToken: gofakeit.UUID(),
		}

		err := repo.CreateTellerLink(context.Background(), &link)
		assert.NoError(t, err, "must be able to create a teller link")
		assert.NotZero(t, link.TellerLinkId, "must now have the ID set")
		assert.Equal(t, user.AccountId, link.AccountId, "should have set the account Id")
	})

}
