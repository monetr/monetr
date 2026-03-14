package plaid_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type DeactivateLinksArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
	LinkId    models.ID[models.Link]    `json:"linkId"`
}

func DeactivatePlaidLinkCron(ctx queue.Context) error {
	log := ctx.Log()

	if !ctx.Configuration().Stripe.IsBillingEnabled() {
		log.DebugContext(ctx, "billing is not enabled, plaid links will not automatically be deactivated")
		crumbs.Debug(ctx, "Billing is not enabled, plaid links will not automatically be deactivated", nil)
		return nil
	}

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())

	log.InfoContext(ctx, "retrieving links for expired accounts")
	expiredLinks, err := jobRepo.GetLinksForExpiredAccounts(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve funding schedules to process")
	}

	if len(expiredLinks) == 0 {
		crumbs.Debug(ctx, "No expired links were found at this time, no work to be done.", nil)
		log.InfoContext(ctx, "no expired links were found at this time, no work to be done")
		return nil
	}

	log.InfoContext(ctx, "preparing to enqueue jobs to remove expired links", "count", len(expiredLinks))

	for _, item := range expiredLinks {
		itemLog := log.With(
			"accountId", item.AccountId,
			"linkId", item.LinkId,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing job to remove expired link for account")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			DeactivatePlaidLink,
			DeactivateLinksArguments{
				AccountId: item.AccountId,
				LinkId:    item.LinkId,
			},
		); err != nil {
			log.WarnContext(ctx, "failed to enqueue job to remove expired link", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to remove expired link", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued removal of expired link")
	}

	return nil
}

func DeactivatePlaidLink(ctx queue.Context, args DeactivateLinksArguments) error {
	return nil
}
