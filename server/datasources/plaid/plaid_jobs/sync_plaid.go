package plaid_jobs

import (
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/pkg/errors"
)

type SyncPlaidArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
	LinkId    models.ID[models.Link]    `json:"linkId"`
	// Trigger will be "webhook" or "manual" or "command"
	Trigger string `json:"trigger"`
}

func SyncPlaidCron(ctx queue.Context) error {
	log := ctx.Log()

	log.InfoContext(ctx, "retrieving links to sync with Plaid")

	links := make([]models.Link, 0)
	cutoff := ctx.Clock().Now().Add(-48 * time.Hour)
	err := ctx.DB().ModelContext(ctx, &links).
		Join(`INNER JOIN "plaid_links" AS "plaid_link"`).
		JoinOn(`"plaid_link"."plaid_link_id" = "link"."plaid_link_id"`).
		Where(`"plaid_link"."status" = ?`, models.PlaidLinkStatusSetup).
		Where(`"plaid_link"."last_attempted_update" < ?`, cutoff).
		Where(`"plaid_link"."deleted_at" IS NULL`).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Where(`"link"."deleted_at" IS NULL`).
		Select(&links)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve links that need to by synced with plaid")
	}

	if len(links) == 0 {
		log.DebugContext(ctx, "no plaid links need to be synced at this time")
		return nil
	}

	log.InfoContext(ctx, "syncing plaid links", "count", len(links))

	for _, item := range links {
		itemLog := log.With(
			"accountId", item.AccountId,
			"linkId", item.LinkId,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing link to be synced with plaid")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			SyncPlaid,
			SyncPlaidArguments{
				AccountId: item.AccountId,
				LinkId:    item.LinkId,
				Trigger:   "cron",
			},
		); err != nil {
			itemLog.WarnContext(ctx, "failed to enqueue job to sync with plaid", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to sync with plaid", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued link to be synced with plaid")
	}

	return nil
}

func SyncPlaid(ctx queue.Context, args SyncPlaidArguments) error {
	return nil
}
