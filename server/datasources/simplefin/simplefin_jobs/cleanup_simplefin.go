package simplefin_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
)

type CleanupSimpleFinArguments struct {
	AccountId       models.ID[models.Account]       `json:"accountId"`
	SimpleFinLinkId models.ID[models.SimpleFinLink] `json:"simpleFinLinkId"`
}

func CleanupSimpleFinCron(ctx queue.Context) error {
	log := ctx.Log()

	if !ctx.Configuration().SimpleFin.IsEnabled() {
		log.InfoContext(ctx, "simplefin is not enabled")
		return nil
	}

	log.InfoContext(ctx, "retrieving stale SimpleFIN links to cleanup")

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())
	staleLinks, err := jobRepo.GetStaleSimpleFinLinks(ctx)
	if err != nil {
		return err
	}

	if len(staleLinks) == 0 {
		log.InfoContext(ctx, "no stale SimpleFIN links to be cleaned up")
		return nil
	}

	if err = queue.BulkEnqueue(
		ctx,
		ctx.Enqueuer(),
		CleanupSimpleFin,
		myownsanity.Map(
			staleLinks,
			func(item models.SimpleFinLink) CleanupSimpleFinArguments {
				return CleanupSimpleFinArguments{
					AccountId:       item.AccountId,
					SimpleFinLinkId: item.SimpleFinLinkId,
				}
			}),
	); err != nil {
		log.WarnContext(ctx, "failed to enqueue jobs to cleanup SimpleFIN links", "err", err)
		crumbs.Warn(ctx, "Failed to enqueue jobs to cleanup SimpleFIN links", "job", map[string]any{
			"error": err,
		})
		return err
	}

	return nil
}

func CleanupSimpleFin(ctx queue.Context, args CleanupSimpleFinArguments) error {
	if !ctx.Configuration().SimpleFin.IsEnabled() {
		ctx.Log().InfoContext(ctx, "simplefin is not enabled")
		return nil
	}

	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		crumbs.IncludeUserInScope(ctx, args.AccountId)
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"simpleFinLinkId", args.SimpleFinLinkId,
		)

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_simplefin",
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

		simpleFinLink, err := repo.GetSimpleFinLink(ctx, args.SimpleFinLinkId)
		if err != nil {
			return err
		}

		if simpleFinLink.Status != models.SimpleFinLinkStatusPending {
			log.WarnContext(ctx, "SimpleFIN link is not in a pending status! There is a bug, please report this!", "bug", true)
			return nil
		}

		// The link has to go before the secret. The secret cascades to the link,
		// but the monetr link is restricted on the SimpleFIN link so deleting the
		// secret first could fail.
		if err := repo.RemoveSimpleFinLink(ctx, simpleFinLink.SimpleFinLinkId); err != nil {
			log.ErrorContext(ctx, "failed to remove SimpleFIN link!", "err", err)
			return err
		}

		if err := secretsRepo.Delete(ctx, simpleFinLink.SecretId); err != nil {
			log.ErrorContext(ctx, "failed to remove SimpleFIN link secret!",
				"err", err,
				"secretId", simpleFinLink.SecretId,
			)
			return err
		}

		return nil
	})
}
