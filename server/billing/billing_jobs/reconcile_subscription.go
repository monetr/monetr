package billing_jobs

import (
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v87"
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
	err := ctx.DB().NewSelect().Model(&accounts).
		Where(`"account"."stripe_customer_id" IS NOT NULL`).
		Where(`"account"."stripe_subscription_id" IS NOT NULL`).
		Where(`"account"."subscription_active_until" < now()`).
		Where(`"account"."subscription_status" = ?`, stripe.SubscriptionStatusActive).
		Where(`"account"."stripe_webhook_latest_timestamp" < ?`, cutoff).
		Limit(100).
		Scan(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to query accounts who may have missed stripe webhooks")
	}

	if len(accounts) == 0 {
		log.InfoContext(ctx, "no accounts have missed webhooks, no subscriptions to reconcile")
		return nil
	}

	log.InfoContext(ctx, "accounts have missed stripe webhooks, subscriptions need to be reconciled", "count", len(accounts))

	if err = queue.BulkEnqueue(
		ctx,
		ctx.Enqueuer(),
		ReconcileSubscription,
		myownsanity.Map(
			accounts,
			func(item models.Account) ReconcileSubscriptionArguments {
				return ReconcileSubscriptionArguments{
					AccountId: item.AccountId,
				}
			}),
	); err != nil {
		log.WarnContext(ctx, "failed to enqueue jobs to reconcile subscriptions", "err", err)
		crumbs.Warn(ctx, "Failed to enqueue jobs to reconcile subscriptions", "job", map[string]any{
			"error": err,
		})
		return err
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
