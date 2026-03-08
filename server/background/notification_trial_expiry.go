package background

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

const (
	NotificationTrialExpiry = "NotificationTrialExpiry"
)

type (
	NotificationTrialExpiryHandler struct {
		log          *slog.Logger
		db           *pg.DB
		clock        clock.Clock
		config       config.Configuration
		email        communication.EmailCommunication
		unmarshaller JobUnmarshaller
	}

	NotificationTrialExpiryArguments struct {
		AccountId ID[Account]
	}

	NotificationTrialExpiryJob struct {
		args   NotificationTrialExpiryArguments
		log    *slog.Logger
		repo   repository.BaseRepository
		db     pg.DBI
		email  communication.EmailCommunication
		clock  clock.Clock
		config config.Configuration
	}
)

func NewNotificationTrialExpiryHandler(
	log *slog.Logger,
	db *pg.DB,
	clock clock.Clock,
	config config.Configuration,
	email communication.EmailCommunication,
) *NotificationTrialExpiryHandler {
	return &NotificationTrialExpiryHandler{
		log:          log,
		db:           db,
		clock:        clock,
		config:       config,
		email:        email,
		unmarshaller: DefaultJobUnmarshaller,
	}
}

func (h NotificationTrialExpiryHandler) QueueName() string {
	return NotificationTrialExpiry
}

func (h NotificationTrialExpiryHandler) DefaultSchedule() string {
	// Run every 6 hours, 30 minutes after the hour.
	return "0 30 */6 * * *"
}

func (h *NotificationTrialExpiryHandler) EnqueueTriggeredJob(
	ctx context.Context,
	enqueuer JobEnqueuer,
) error {
	log := h.log

	var accounts []Account
	cutoff := h.clock.Now().AddDate(0, 0, 5)

	err := h.db.ModelContext(ctx, &accounts).
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
		err := enqueuer.EnqueueJob(ctx, h.QueueName(), NotificationTrialExpiryArguments{
			AccountId: item.AccountId,
		})
		if err != nil {
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

func (h *NotificationTrialExpiryHandler) HandleConsumeJob(
	ctx context.Context,
	log *slog.Logger,
	data []byte,
) error {
	var args NotificationTrialExpiryArguments
	if err := errors.Wrap(h.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return h.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log = log.With("accountId", args.AccountId)

		repo := repository.NewRepositoryFromSession(
			h.clock,
			"user_system",
			args.AccountId,
			txn,
			log,
		)
		job, err := NewNotificationTrialExpiryJob(
			log,
			repo,
			txn,
			h.clock,
			h.config,
			h.email,
			args,
		)
		if err != nil {
			return err
		}
		return job.Run(ctx)
	})
}

func NewNotificationTrialExpiryJob(
	log *slog.Logger,
	repo repository.BaseRepository,
	db pg.DBI,
	clock clock.Clock,
	config config.Configuration,
	email communication.EmailCommunication,
	args NotificationTrialExpiryArguments,
) (*NotificationTrialExpiryJob, error) {
	return &NotificationTrialExpiryJob{
		args:   args,
		log:    log,
		repo:   repo,
		db:     db,
		email:  email,
		clock:  clock,
		config: config,
	}, nil
}

func (j *NotificationTrialExpiryJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := j.log.With("accountId", j.args.AccountId)

	log.Log(span.Context(), logging.LevelTrace, "looking up account owner")
	owner, err := j.repo.GetAccountOwner(span.Context())
	if err != nil {
		return err
	}

	log = log.With(
		"loginId", owner.LoginId,
		"userId", owner.UserId,
	)

	if owner.Account.TrialExpiryNotificationSentAt != nil {
		log.DebugContext(span.Context(), "notification has already been sent for account")
		return nil
	}

	if j.config.Email.Verification.Enabled && owner.Login.EmailVerifiedAt == nil {
		log.InfoContext(span.Context(), "skipping trial expiry notification, owner has not verified their email address")
		return nil
	}

	myownsanity.ASSERT_NOTNIL(owner.Account.TrialEndsAt, "trial ends at must be present for trial notifications")

	now := j.clock.Now()
	owner.Account.TrialExpiryNotificationSentAt = &now

	// TODO This is really ugly, ideally this would be done via the account
	// repository. But maybe this field shouldn't even be on that object? Maybe it
	// should be in its own table somehow? Maybe in the long run the billing
	// interface should handle this expiry email notification as well, then we
	// don't need to prop-drill the account repo everywhere?
	_, err = j.db.ModelContext(span.Context(), owner.Account).
		WherePK().
		Update(owner.Account)
	if err != nil {
		return errors.Wrap(err, "failed to mark account as notified about trial expiration")
	}

	expiration := *owner.Account.TrialEndsAt

	timezone, err := owner.Account.GetTimezone()
	if err != nil {
		log.WarnContext(span.Context(), "failed to get timezone for account trial notification", "err", err)
		timezone = time.UTC
	}

	expiration = expiration.In(timezone)
	now = now.In(timezone)

	days := int(expiration.Sub(now).Hours() / 24)
	email := communication.TrialAboutToExpireParams{
		BaseURL:               j.config.Server.GetBaseURL().String(),
		Email:                 owner.Login.Email,
		FirstName:             owner.Login.FirstName,
		LastName:              owner.Login.LastName,
		TrialExpirationDate:   expiration.Format("Monday January 2, 2006"),
		TrialExpirationWindow: fmt.Sprintf("%d days", days),
		SupportEmail:          "support@monetr.app",
	}

	log.InfoContext(span.Context(), "sending trial expiry notification")

	if err = j.email.SendEmail(span.Context(), email); err != nil {
		return errors.Wrap(err, "failed to send trial expiry notification email")
	}

	return nil
}
