package billing_jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/billing/billing_jobs"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_stripe"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v81"
	"go.uber.org/mock/gomock"
)

func TestNotificationTrialExpiryCron(t *testing.T) {
	t.Run("notify users before trial expires", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		{ // Mark the login's email as verified
			login.IsEmailVerified = true
			login.EmailVerifiedAt = myownsanity.Pointer(clock.Now())
			testutils.MustDBUpdate(t, &login)
		}
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{
			Stripe: config.Stripe{
				Enabled: true,
			},
		}
		email := mockgen.NewMockEmailCommunication(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			err := billing_jobs.NotificationTrialExpiryCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}

		// Move time forward 26 days, the trial in the fixture is for 30 days; so we
		// should now be within the notification window.
		clock.Add(26 * 24 * time.Hour)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(billing_jobs.NotificationTrialExpiry),
				gomock.Any(),
				gomock.Eq(billing_jobs.NotificationTrialExpiryArguments{
					AccountId: user.AccountId,
				}),
			).
			Return(nil).
			Times(1)

		// Run the cron again, this time we should see the notification get
		// enqueued.
		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			err := billing_jobs.NotificationTrialExpiryCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}

		// Now update the account to show that the notification has been sent.
		user.Account.TrialExpiryNotificationSentAt = myownsanity.Pointer(clock.Now())
		testutils.MustDBUpdate(t, user.Account)

		// And run the cron again, this time we should not do anything because the
		// notification will have already been sent.
		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			err := billing_jobs.NotificationTrialExpiryCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("dont notify users who have not verified their email", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		{ // Mark the login's email as not verified
			login.IsEmailVerified = false
			login.EmailVerifiedAt = nil
			testutils.MustDBUpdate(t, &login)
		}
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{
			Stripe: config.Stripe{
				Enabled: true,
			},
		}
		email := mockgen.NewMockEmailCommunication(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			// First time, no notifications should be enqueued and we should not have
			// an error.
			err := billing_jobs.NotificationTrialExpiryCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}

		// Move time forward 26 days, the trial in the fixture is for 30 days; so we
		// should now be within the notification window.
		clock.Add(26 * 24 * time.Hour)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(billing_jobs.NotificationTrialExpiry),
				gomock.Any(),
				gomock.Eq(billing_jobs.NotificationTrialExpiryArguments{
					AccountId: user.AccountId,
				}),
			).
			Return(nil).
			Times(0)

		// Run the cron again, nothing should happen because the user in this test
		// has not verified their email
		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			err := billing_jobs.NotificationTrialExpiryCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("user subscribed before trial ended", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{
			Stripe: config.Stripe{
				Enabled: true,
			},
		}
		email := mockgen.NewMockEmailCommunication(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		clock.Add(24 * time.Hour)

		// After 24 hours the user decided to subscribe early. Update the account.
		status := stripe.SubscriptionStatusActive
		user.Account.SubscriptionActiveUntil = myownsanity.Pointer(clock.Now().AddDate(0, 0, 30))
		user.Account.SubscriptionStatus = &status
		user.Account.StripeCustomerId = myownsanity.Pointer(mock_stripe.FakeStripeCustomerId(t))
		user.Account.StripeSubscriptionId = myownsanity.Pointer(mock_stripe.FakeStripeSubscriptionId(t))
		testutils.MustDBUpdate(t, user.Account)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			// First time, no notifications should be enqueued and we should not have
			// an error.
			err := billing_jobs.NotificationTrialExpiryCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}

		// Make sure that we don't send an email because the user is already
		// subscribed.
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(billing_jobs.NotificationTrialExpiry),
				gomock.Any(),
				gomock.Eq(billing_jobs.NotificationTrialExpiryArguments{
					AccountId: user.AccountId,
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
		config := config.Configuration{
			Stripe: config.Stripe{
				Enabled: true,
			},
		}
		email := mockgen.NewMockEmailCommunication(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		// Move us 90 days into the future, well after the trial has ended.
		clock.Add(90 * 24 * time.Hour)

		// Setup the user as subscribed at the 90 day mark.
		status := stripe.SubscriptionStatusActive
		user.Account.SubscriptionActiveUntil = myownsanity.Pointer(clock.Now().AddDate(0, 0, 30))
		user.Account.SubscriptionStatus = &status
		user.Account.StripeCustomerId = myownsanity.Pointer(mock_stripe.FakeStripeCustomerId(t))
		user.Account.StripeSubscriptionId = myownsanity.Pointer(mock_stripe.FakeStripeSubscriptionId(t))
		testutils.MustDBUpdate(t, user.Account)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			// First time, no notifications should be enqueued and we should not have
			// an error.
			err := billing_jobs.NotificationTrialExpiryCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}

		// Make sure that we don't send an email because the user is already
		// subscribed.
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(billing_jobs.NotificationTrialExpiry),
				gomock.Any(),
				gomock.Eq(billing_jobs.NotificationTrialExpiryArguments{
					AccountId: user.AccountId,
				}),
			).
			Return(nil).
			Times(0)
	})
}

func TestNotificationTrialExpiry(t *testing.T) {
	t.Run("will send notification", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{
			Stripe: config.Stripe{
				Enabled: true,
			},
		}
		email := mockgen.NewMockEmailCommunication(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

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

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := billing_jobs.NotificationTrialExpiry(
				mockqueue.NewMockContext(context),
				billing_jobs.NotificationTrialExpiryArguments{
					AccountId: user.AccountId,
				},
			)
			assert.NoError(t, err)
		}

		updatedAccount := testutils.MustDBRead(t, *user.Account)
		assert.NotNil(t, updatedAccount.TrialExpiryNotificationSentAt, "notification timestamp should be set now")
	})

	t.Run("already sent a notification", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		user := fixtures.GivenIHaveATrialingAccount(t, clock, login)
		config := config.Configuration{
			Stripe: config.Stripe{
				Enabled: true,
			},
		}
		email := mockgen.NewMockEmailCommunication(ctrl)

		assert.NotNil(t, user.Account, "need account to force trial scenario")
		assert.NotNil(t, user.Account.TrialEndsAt, "trial ends at date must be present")
		assert.Nil(t, user.Account.TrialExpiryNotificationSentAt, "trial notification email should be unsent")

		// Set the timestamp as if we have already sent the notification.
		user.Account.TrialExpiryNotificationSentAt = myownsanity.Pointer(clock.Now())
		testutils.MustDBUpdate(t, user.Account)

		email.EXPECT().
			SendEmail(
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Email().Return(email).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := billing_jobs.NotificationTrialExpiry(
				mockqueue.NewMockContext(context),
				billing_jobs.NotificationTrialExpiryArguments{
					AccountId: user.AccountId,
				},
			)
			assert.NoError(t, err)
		}
	})
}
