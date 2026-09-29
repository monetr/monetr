package lunch_flow_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
)

type CleanupLunchFlowArguments struct {
	AccountId       models.ID[models.Account]       `json:"accountId"`
	LunchFlowLinkId models.ID[models.LunchFlowLink] `json:"lunchFlowLinkId"`
}

func CleanupLunchFlowCron(ctx queue.Context) error {
	log := ctx.Log()

	if !ctx.Configuration().LunchFlow.Enabled {
		log.InfoContext(ctx, "lunch flow is not enabled")
		return nil
	}

	log.InfoContext(ctx, "retrieving bank accounts to sync with Lunch Flow")

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())
	staleLinks, err := jobRepo.GetStaleLunchFlowLinks(ctx)
	if err != nil {
		return err
	}

	if len(staleLinks) == 0 {
		log.InfoContext(ctx, "no stale Lunch Flow links to be cleaned up")
		return nil
	}

	if err = queue.BulkEnqueue(
		ctx,
		ctx.Enqueuer(),
		CleanupLunchFlow,
		myownsanity.Map(
			staleLinks,
			func(item models.LunchFlowLink) CleanupLunchFlowArguments {
				return CleanupLunchFlowArguments{
					AccountId:       item.AccountId,
					LunchFlowLinkId: item.LunchFlowLinkId,
				}
			}),
	); err != nil {
		log.WarnContext(ctx, "failed to enqueue jobs to cleanup Lunch Flow links", "err", err)
		crumbs.Warn(ctx, "Failed to enqueue jobs to cleanup Lunch Flow links", "job", map[string]any{
			"error": err,
		})
		return err
	}

	return nil
}

func CleanupLunchFlow(ctx queue.Context, args CleanupLunchFlowArguments) error {
	if !ctx.Configuration().LunchFlow.Enabled {
		ctx.Log().InfoContext(ctx, "lunch flow is not enabled")
		return nil
	}

	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		crumbs.IncludeUserInScope(ctx, args.AccountId)
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"lunchFlowLinkId", args.LunchFlowLinkId,
		)

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_lunch_flow",
			args.AccountId,
			ctx.DB(),
			log,
		)
		secretsRepo := repository.NewSecretsRepository(
			log,
			ctx.Clock(),
			ctx.DB(),
			ctx.KMS(),
			args.AccountId,
		)

		lunchFlowLink, err := repo.GetLunchFlowLink(ctx, args.LunchFlowLinkId)
		if err != nil {
			return err
		}

		log = log.With("lunchFlowLinkId", lunchFlowLink.LunchFlowLinkId)

		if lunchFlowLink.Status != models.LunchFlowLinkStatusPending {
			log.WarnContext(ctx, "Lunch Flow link is not in a pending status! There is a bug, please report this!", "bug", true)
			return nil
		}

		// Because we cascade deletes for lunch flow links, we do not need additional
		// complexity or steps here. We just need to delete child objects that won't
		// get cascaded, such as secrets.
		if err := repo.RemoveLunchFlowLink(ctx, lunchFlowLink.LunchFlowLinkId); err != nil {
			log.ErrorContext(ctx, "failed to remove Lunch Flow link!", "err", err)
			return err
		}

		if err := secretsRepo.Delete(ctx, lunchFlowLink.SecretId); err != nil {
			log.ErrorContext(ctx, "failed to remove Lunch Flow link secret!",
				"err", err,
				"secretId", lunchFlowLink.SecretId,
			)
			return err
		}

		return nil
	})
}
