package billing_jobs

import (
	"fmt"
	"time"

	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type NotificationTrialExpiryArguments struct {
	AccountId models.ID[models.Account]
}

func NotificationTrialExpiryCron(ctx queue.Context) error {
	log := ctx.Log()

	if !ctx.Configuration().Stripe.IsBillingEnabled() {
		log.DebugContext(ctx, "billing is not enabled, no trial notifications necessary")
		crumbs.Debug(ctx, "Billing is not enabled, no trial notifications necessary", nil)
		return nil
	}

	var accounts []models.Account
	cutoff := ctx.Clock().Now().AddDate(0, 0, 5)

	err := ctx.DB().ModelContext(ctx, &accounts).
		Join(`INNER JOIN "users" AS "user"`).
		JoinOn(`"user"."account_id" = "account"."account_id"`).
		Join(`INNER JOIN "logins" AS "login"`).
		JoinOn(`"login"."login_id" = "user"."login_id"`).
		Where(`"account"."trial_expiry_notification_sent_at" IS NULL`).
		Where(`"account"."trial_ends_at" < ?`, cutoff).
		// Make sure to exclude users who have subscribed before their trial ends.
		Where(`"account"."subscription_active_until" IS NULL`).
		// Only if there is an owner on the account.
		Where(`"user"."role" = ?`, models.UserRoleOwner).
		// Only if the owner has verified.
		Where(`"login"."email_verified_at" IS NOT NULL`).
		Select(&accounts)
	if err != nil {
		return errors.Wrap(err, "failed to query accounts who need a trial expiry notification")
	}

	if len(accounts) == 0 {
		log.InfoContext(ctx, "no accounts need a trial expiry notification at this time")
		return nil
	}

	log.InfoContext(ctx, "accounts need a trial expiry notification", "count", len(accounts))

	for _, item := range accounts {
		itemLog := log.With("accountId", item.AccountId)

		itemLog.Log(ctx, logging.LevelTrace, "enqueuing account for trial expiry notification")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			NotificationTrialExpiry,
			NotificationTrialExpiryArguments{
				AccountId: item.AccountId,
			},
		); err != nil {

			itemLog.WarnContext(ctx, "failed to enqueue job for trial expiry notification", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job for trial expiry notification", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued account for trial expiry notification")
	}

	return nil
}

func NotificationTrialExpiry(ctx queue.Context, args NotificationTrialExpiryArguments) error {
	log := ctx.Log().With("accountId", args.AccountId)

	if !ctx.Configuration().Stripe.IsBillingEnabled() {
		log.DebugContext(ctx, "billing is not enabled, no trial notifications necessary")
		crumbs.Debug(ctx, "Billing is not enabled, no trial notifications necessary", nil)
		return nil
	}

	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_system",
			args.AccountId,
			ctx.DB(),
			log,
		)

		log.Log(ctx, logging.LevelTrace, "looking up account owner")
		owner, err := repo.GetAccountOwner(ctx)
		if err != nil {
			return err
		}

		log = log.With(
			"loginId", owner.LoginId,
			"userId", owner.UserId,
		)

		if owner.Account.TrialExpiryNotificationSentAt != nil {
			log.DebugContext(ctx, "notification has already been sent for account")
			return nil
		}

		if ctx.Configuration().Email.Verification.Enabled && owner.Login.EmailVerifiedAt == nil {
			log.InfoContext(ctx, "skipping trial expiry notification, owner has not verified their email address")
			return nil
		}

		myownsanity.ASSERT_NOTNIL(
			owner.Account.TrialEndsAt,
			"trial ends at must be present for trial notifications",
		)

		now := ctx.Clock().Now()
		owner.Account.TrialExpiryNotificationSentAt = &now

		// TODO This is really ugly, ideally this would be done via the account
		// repository. But maybe this field shouldn't even be on that object? Maybe it
		// should be in its own table somehow? Maybe in the long run the billing
		// interface should handle this expiry email notification as well, then we
		// don't need to prop-drill the account repo everywhere?
		_, err = ctx.DB().ModelContext(ctx, owner.Account).
			WherePK().
			Update(owner.Account)
		if err != nil {
			return errors.Wrap(err, "failed to mark account as notified about trial expiration")
		}

		expiration := *owner.Account.TrialEndsAt

		timezone, err := owner.Account.GetTimezone()
		if err != nil {
			log.WarnContext(ctx, "failed to get timezone for account trial notification", "err", err)
			timezone = time.UTC
		}

		expiration = expiration.In(timezone)
		now = now.In(timezone)

		days := int(expiration.Sub(now).Hours() / 24)
		email := communication.TrialAboutToExpireParams{
			BaseURL:               ctx.Configuration().Server.GetBaseURL().String(),
			Email:                 owner.Login.Email,
			FirstName:             owner.Login.FirstName,
			LastName:              owner.Login.LastName,
			TrialExpirationDate:   expiration.Format("Monday January 2, 2006"),
			TrialExpirationWindow: fmt.Sprintf("%d days", days),
			SupportEmail:          "support@monetr.app",
		}

		log.InfoContext(ctx, "sending trial expiry notification")

		if err = ctx.Email().SendEmail(ctx, email); err != nil {
			return errors.Wrap(err, "failed to send trial expiry notification email")
		}

		return nil
	})
}
