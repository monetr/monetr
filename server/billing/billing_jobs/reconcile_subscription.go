package billing_jobs

import (
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v81"
)

type ReconcileSubscriptionArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
}

func ReconcileSubscriptionCron(ctx queue.Context) error {
	if !ctx.Configuration().Stripe.IsBillingEnabled() {
		ctx.Log().DebugContext(ctx, "billing is not enabled, no reconcile necesssary")
		crumbs.Debug(ctx, "Billing is not enabled, no recocile necessary", nil)
		return nil
	}

	log := ctx.Log()

	var accounts []models.Account
	cutoff := ctx.Clock().Now().Add(-12 * time.Hour)
	err := ctx.DB().ModelContext(ctx, &accounts).
		Where(`"account"."stripe_customer_id" IS NOT NULL`).
		Where(`"account"."stripe_subscription_id" IS NOT NULL`).
		Where(`"account"."subscription_active_until" < now()`).
		Where(`"account"."subscription_status" = ?`, stripe.SubscriptionStatusActive).
		Where(`"account"."stripe_webhook_latest_timestamp" < ?`, cutoff).
		Limit(100).
		Select(&accounts)
	if err != nil {
		return errors.Wrap(err, "failed to query accounts who may have missed stripe webhooks")
	}

	if len(accounts) == 0 {
		log.InfoContext(ctx, "no accounts have missed webhooks, no subscriptions to reconcile")
		return nil
	}

	log.InfoContext(ctx, "accounts have missed stripe webhooks, subscriptions need to be reconciled", "count", len(accounts))

	for _, item := range accounts {
		itemLog := log.With("accountId", item.AccountId)

		itemLog.Log(ctx, logging.LevelTrace, "enqueuing account to have subscription reconciled")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			ReconcileSubscription,
			ReconcileSubscriptionArguments{
				AccountId: item.AccountId,
			},
		); err != nil {

			itemLog.WarnContext(ctx, "failed to enqueue job to reconcile subscription", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to reconcile subscription", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued account for subscription reconciliation")
	}

	return nil
}

func ReconcileSubscription(ctx queue.Context, args ReconcileSubscriptionArguments) error {
	if !ctx.Configuration().Stripe.IsBillingEnabled() {
		ctx.Log().DebugContext(ctx, "billing is not enabled, no reconcile necesssary")
		crumbs.Debug(ctx, "Billing is not enabled, no recocile necessary", nil)
		return nil
	}
	return ctx.Billing().ReconcileSubscription(ctx, args.AccountId)
}
