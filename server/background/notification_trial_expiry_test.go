package background_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/repository"
	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNotificationTrialExpiryHandler_DefaultSchedule(t *testing.T) {
	t.Run("validate cron schedule", func(t *testing.T) {
		handler := &background.NotificationTrialExpiryHandler{}
		schedule, err := cron.Parse(handler.DefaultSchedule())
		assert.NoError(t, err, "must be able too parse the schedule")
		now := time.Now()
		next := schedule.Next(now)
		assert.GreaterOrEqual(t, next, now, "next cron should always be greater or equal than now")
	})
}

func TestNotificationTrialExpiryHandler_EnqueueTriggeredJob(t *testing.T) {
	t.Run("notify users before trial expires", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{}
		email := mockgen.NewMockEmailCommunication(ctrl)
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		handler := background.NewNotificationTrialExpiryHandler(
			log,
			db,
			clock,
			config,
			email,
		)

		// First time, no notifications should be enqueued and we should not have an
		// error.
		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)

		// Move time forward 26 days, the trial in the fixture is for 30 days; so we
		// should now be within the notification window.
		clock.Add(26 * 24 * time.Hour)

		// But now we should see one call to enqueue the notification.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.NotificationTrialExpiry),
				testutils.NewGenericMatcher(func(args background.NotificationTrialExpiryArguments) bool {
					return assert.EqualValues(t, user.AccountId, args.AccountId)
				}),
			).
			Return(nil).
			Times(1)

		// Run the cron again, this time we should see the notification get
		// enqueued.
		err = handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)

		// Now update the account to show that the notification has been sent.
		user.Account.TrialExpiryNotificationSentAt = myownsanity.TimeP(clock.Now())
		testutils.MustDBUpdate(t, user.Account)

		// And run the cron again, this time we should not do anything because the
		// notification will have already been sent.
		err = handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)
	})
}

func TestNotificationTrialExpiryJob_Run(t *testing.T) {
	t.Run("will send notification", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{}
		email := mockgen.NewMockEmailCommunication(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		job, err := background.NewNotificationTrialExpiryJob(
			log,
			repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db),
			db,
			clock,
			config,
			email,
			background.NotificationTrialExpiryArguments{
				AccountId: user.AccountId,
			},
		)
		assert.NoError(t, err, "must not return an error just creating the job")

		email.EXPECT().
			SendEmail(
				gomock.Any(),
				gomock.AssignableToTypeOf(communication.TrialAboutToExpireParams{}),
			).
			Return(nil).
			Times(1).
			Do(func(ctx context.Context, params communication.TrialAboutToExpireParams) error {
				assert.Equal(t, user.Login.Email, params.Email)
				assert.Equal(t, user.Login.FirstName, params.FirstName)
				assert.Equal(t, user.Login.LastName, params.LastName)
				assert.NotEmpty(t, params.TrialExpirationWindow)
				assert.NotEmpty(t, params.TrialExpirationDate)
				return nil
			})

		err = job.Run(context.Background())
		assert.NoError(t, err)

		updatedAccount := testutils.MustDBRead(t, *user.Account)
		assert.NotNil(t, updatedAccount.TrialExpiryNotificationSentAt, "notification timestamp should be set now")
	})

	t.Run("already sent a notification", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{}
		email := mockgen.NewMockEmailCommunication(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		// Set the timestamp as if we have already sent the notification.
		user.Account.TrialExpiryNotificationSentAt = myownsanity.TimeP(clock.Now())
		testutils.MustDBUpdate(t, user.Account)

		job, err := background.NewNotificationTrialExpiryJob(
			log,
			repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db),
			db,
			clock,
			config,
			email,
			background.NotificationTrialExpiryArguments{
				AccountId: user.AccountId,
			},
		)
		assert.NoError(t, err, "must not return an error just creating the job")

		email.EXPECT().
			SendEmail(
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		err = job.Run(context.Background())
		assert.NoError(t, err)
	})
}
