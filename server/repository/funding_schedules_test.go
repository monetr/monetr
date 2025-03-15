package repository_test

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_UpdateFundingSchedule(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=DAILY", false)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		fundingSchedule.Name = "Updated name"

		err := repo.UpdateFundingSchedule(context.Background(), fundingSchedule)
		assert.NoError(t, err, "must be able to update funding schedule")
		updatedSchedule := testutils.MustDBRead(t, *fundingSchedule)
		assert.Equal(t, "Updated name", updatedSchedule.Name, "name should match the new one")
	})
}
