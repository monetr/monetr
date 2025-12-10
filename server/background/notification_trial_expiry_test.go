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
	"github.com/monetr/monetr/server/internal/mock_stripe"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/repository"
	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v81"
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
		{ // Mark the login's email as verified
			login.IsEmailVerified = true
			login.EmailVerifiedAt = myownsanity.TimeP(clock.Now())
			testutils.MustDBUpdate(t, &login)
		}
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

	t.Run("dont notify users who have not verified their email", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		{ // Mark the login's email as verified
			login.IsEmailVerified = false
			login.EmailVerifiedAt = nil
			testutils.MustDBUpdate(t, &login)
		}
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
			Times(0)

		// Run the cron again, this time we should see the notification get
		// enqueued.
		err = handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)
	})

	t.Run("user subscribed before trial ended", func(t *testing.T) {
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

		handler := background.NewNotificationTrialExpiryHandler(
			log,
			db,
			clock,
			config,
			email,
		)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		clock.Add(24 * time.Hour)

		// After 24 hours the user decided to subscribe early. Update the account.
		status := stripe.SubscriptionStatusActive
		user.Account.SubscriptionActiveUntil = myownsanity.TimeP(clock.Now().AddDate(0, 0, 30))
		user.Account.SubscriptionStatus = &status
		user.Account.StripeCustomerId = myownsanity.StringP(mock_stripe.FakeStripeCustomerId(t))
		user.Account.StripeSubscriptionId = myownsanity.StringP(mock_stripe.FakeStripeSubscriptionId(t))
		testutils.MustDBUpdate(t, user.Account)

		// First time, no notifications should be enqueued and we should not have an
		// error.
		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)

		// Make sure that we don't send an email because the user is already
		// subscribed.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.NotificationTrialExpiry),
				testutils.NewGenericMatcher(func(args background.NotificationTrialExpiryArguments) bool {
					return assert.EqualValues(t, user.AccountId, args.AccountId)
				}),
			).
			Return(nil).
			Times(0)
	})

	t.Run("user subscribed after trial ended", func(t *testing.T) {
		// This test technically should prove a behavior that should never happen.
		// When a user has subscribed _after_ their trial has ended, but we haven't
		// sent them a trial notification email. This way we don't send dumb emails
		// if the job is dead for a while or something.
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

		handler := background.NewNotificationTrialExpiryHandler(
			log,
			db,
			clock,
			config,
			email,
		)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		// Move us 90 days into the future, well after the trial has ended.
		clock.Add(90 * 24 * time.Hour)

		// Setup the user as subscribed at the 90 day mark.
		status := stripe.SubscriptionStatusActive
		user.Account.SubscriptionActiveUntil = myownsanity.TimeP(clock.Now().AddDate(0, 0, 30))
		user.Account.SubscriptionStatus = &status
		user.Account.StripeCustomerId = myownsanity.Pointer(mock_stripe.FakeStripeCustomerId(t))
		user.Account.StripeSubscriptionId = myownsanity.Pointer(mock_stripe.FakeStripeSubscriptionId(t))
		testutils.MustDBUpdate(t, user.Account)

		// First time, no notifications should be enqueued and we should not have an
		// error.
		err := handler.EnqueueTriggeredJob(context.Background(), enqueuer)
		assert.NoError(t, err)

		// Make sure that we don't send an email because the user is already
		// subscribed.
		enqueuer.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.NotificationTrialExpiry),
				testutils.NewGenericMatcher(func(args background.NotificationTrialExpiryArguments) bool {
					return assert.EqualValues(t, user.AccountId, args.AccountId)
				}),
			).
			Return(nil).
			Times(0)
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

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		job, err := background.NewNotificationTrialExpiryJob(
			log,
			repo,
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

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		job, err := background.NewNotificationTrialExpiryJob(
			log,
			repo,
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
