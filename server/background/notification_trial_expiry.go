package background

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	NotificationTrialExpiry = "NotificationTrialExpiry"
)

type (
	NotificationTrialExpiryHandler struct {
		log          *logrus.Entry
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
		log    *logrus.Entry
		repo   repository.BaseRepository
		db     pg.DBI
		email  communication.EmailCommunication
		clock  clock.Clock
		config config.Configuration
	}
)

func NewNotificationTrialExpiryHandler(
	log *logrus.Entry,
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
	log := h.log.WithContext(ctx)

	var accounts []Account
	cutoff := h.clock.Now().AddDate(0, 0, 5)

	err := h.db.ModelContext(ctx, &accounts).
		Where(`"account"."trial_expiry_notification_sent_at" IS NULL`).
		Where(`"account"."trial_ends_at" < ?`, cutoff).
		// Make sure to exclude users who have subscribed before their trial ends.
		Where(`"account"."subscription_active_until" IS NULL`).
		Select(&accounts)
	if err != nil {
		return errors.Wrap(err, "failed to query accounts who need a trial expiry notification")
	}

	if len(accounts) == 0 {
		log.Info("no accounts need a trial expiry notification at this time")
		return nil
	}

	log.WithField("count", len(accounts)).Info("accounts need a trial expiry notification")

	for _, item := range accounts {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
		})

		itemLog.Trace("enqueuing account for trial expiry notification")
		err := enqueuer.EnqueueJob(ctx, h.QueueName(), NotificationTrialExpiryArguments{
			AccountId: item.AccountId,
		})
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job for trial expiry notification")
			crumbs.Warn(ctx, "Failed to enqueue job for trial expiry notification", "job", map[string]interface{}{
				"error": err,
			})
			continue
		}

		itemLog.Trace("successfully enqueued account for trial expiry notification")
	}

	return nil
}

func (h *NotificationTrialExpiryHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args NotificationTrialExpiryArguments
	if err := errors.Wrap(h.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return h.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		repo := repository.NewRepositoryFromSession(
			h.clock,
			"user_system",
			args.AccountId,
			txn,
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
	log *logrus.Entry,
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

	log := j.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"accountId": j.args.AccountId,
	})

	log.Trace("looking up account owner")
	owner, err := j.repo.GetAccountOwner(span.Context())
	if err != nil {
		return err
	}

	log = log.WithFields(logrus.Fields{
		"loginId": owner.LoginId,
		"userId":  owner.UserId,
	})

	if owner.Account.TrialExpiryNotificationSentAt != nil {
		log.Debug("notification has already been sent for account")
		return nil
	}

	if owner.Login.EmailVerifiedAt == nil {
		log.Info("skipping trial expiry notification, owner has not verified their email address")
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
		log.WithError(err).Warn("failed to get timezone for account trial notification")
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

	log.Info("sending trial expiry notification")

	if err = j.email.SendEmail(span.Context(), email); err != nil {
		return errors.Wrap(err, "failed to send trial expiry notification email")
	}

	return nil
}
