package background

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81"
)

const (
	ReconcileSubscription = "ReconcileSubscription"
)

type (
	ReconcileSubscriptionHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		publisher    pubsub.Publisher
		billing      billing.Billing
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	ReconcileSubscriptionArguments struct {
		AccountId ID[Account] `json:"accountId"`
	}

	ReconcileSubscriptionJob struct {
		args      ReconcileSubscriptionArguments
		log       *logrus.Entry
		repo      repository.BaseRepository
		publisher pubsub.Publisher
		billing   billing.Billing
		clock     clock.Clock
	}
)

func TriggerReconcileSubscription(
	ctx context.Context,
	backBackgroundJobs JobController,
	arguments ReconcileSubscriptionArguments,
) error {
	return backBackgroundJobs.EnqueueJob(ctx, ReconcileSubscription, arguments)
}

func NewReconcileSubscriptionHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	publisher pubsub.Publisher,
	billing billing.Billing,
) *ReconcileSubscriptionHandler {
	return &ReconcileSubscriptionHandler{
		log:          log,
		db:           db,
		publisher:    publisher,
		billing:      billing,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (h ReconcileSubscriptionHandler) QueueName() string {
	return ReconcileSubscription
}

func (h *ReconcileSubscriptionHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args ReconcileSubscriptionArguments
	if err := errors.Wrap(h.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for reconcile subscription job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	log = log.WithContext(ctx).WithFields(logrus.Fields{
		"accountId": args.AccountId,
	})

	repo := repository.NewRepositoryFromSession(
		h.clock,
		"user_system",
		args.AccountId,
		h.db,
		log,
	)
	job, err := NewReconcileSubscriptionJob(
		log,
		repo,
		h.clock,
		h.publisher,
		h.billing,
		args,
	)
	if err != nil {
		return err
	}
	return job.Run(ctx)
}

func (h ReconcileSubscriptionHandler) DefaultSchedule() string {
	// Run every 12 hours, 15 minutes after the hour.
	return "0 15 */12 * * *"
}

func (h *ReconcileSubscriptionHandler) EnqueueTriggeredJob(
	ctx context.Context,
	enqueuer JobEnqueuer,
) error {
	log := h.log.WithContext(ctx)

	var accounts []Account
	cutoff := h.clock.Now().Add(-12 * time.Hour)
	err := h.db.ModelContext(ctx, &accounts).
		Where(`"account"."stripe_customer_id" IS NOT NULL`).
		Where(`"account"."stripe_subscription_id" IS NOT NULL`).
		Where(`"account"."subscription_active_until" < now()`).
		Where(`"account"."subscription_status" = ?`, stripe.SubscriptionStatusActive).
		Where(`"account"."stripe_webhook_latest_timestamp" < ?`, cutoff).
		Limit(1000).
		Select(&accounts)
	if err != nil {
		return errors.Wrap(err, "failed to query accounts who may have missed stripe webhooks")
	}

	if len(accounts) == 0 {
		log.Info("no accounts have missed webhooks, no subscriptions to reconcile")
		return nil
	}

	log.WithField("count", len(accounts)).
		Info("accounts have missed stripe webhooks, subscriptions need to be reconciled")

	for _, item := range accounts {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
		})

		itemLog.Trace("enqueuing account to have subscription reconciled")
		err := enqueuer.EnqueueJob(ctx, h.QueueName(), ReconcileSubscriptionArguments{
			AccountId: item.AccountId,
		})
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job to reconcile subscription")
			crumbs.Warn(ctx, "Failed to enqueue job to reconcile subscription", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Trace("successfully enqueued account for subscription reconciliation")
	}

	return nil
}

func NewReconcileSubscriptionJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	publisher pubsub.Publisher,
	billing billing.Billing,
	args ReconcileSubscriptionArguments,
) (*ReconcileSubscriptionJob, error) {
	return &ReconcileSubscriptionJob{
		args:      args,
		log:       log,
		repo:      repo,
		publisher: publisher,
		billing:   billing,
		clock:     clock,
	}, nil
}

func (j *ReconcileSubscriptionJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	return j.billing.ReconcileSubscription(span.Context(), j.args.AccountId)
}
