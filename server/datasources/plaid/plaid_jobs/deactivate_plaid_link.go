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
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		crumbs.IncludeUserInScope(ctx, args.AccountId)
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"linkId", args.LinkId,
		)

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_system",
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

		link, err := repo.GetLink(ctx, args.LinkId)
		if err != nil {
			log.ErrorContext(ctx, "failed to retrieve link for deactivation", "err", err)
			return err
		}

		if link.PlaidLink == nil {
			log.WarnContext(ctx, "provided link does not have any plaid credentials")
			crumbs.Warn(ctx, "BUG: Link was queued to be deactivated, but has no plaid details", "jobs", map[string]any{
				"link": link,
			})
			return nil
		}

		log = log.With(
			"plaidLinkId", link.PlaidLinkId,
			"itemId", link.PlaidLink.PlaidId,
			"institutionId", link.PlaidLink.InstitutionId,
			"institutionName", link.PlaidLink.InstitutionName,
			"plaidLinkStatus", link.PlaidLink.Status,
		)

		log.InfoContext(ctx, "removing plaid link")

		secret, err := secretsRepo.Read(ctx, link.PlaidLink.SecretId)
		if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
			log.ErrorContext(ctx, "could not retrieve API credentials for Plaid for link, this job will be retried", "err", err)
			return err
		}

		client, err := ctx.Platypus().NewClient(ctx, link, secret.Value, link.PlaidLink.PlaidId)
		if err != nil {
			log.ErrorContext(ctx, "failed to create client for link deactivation", "err", err)
			return err
		}

		if err = repo.DeletePlaidLink(ctx, *link.PlaidLinkId); err != nil {
			log.WarnContext(ctx, "failed to remove Plaid details, link cannot be removed at this time", "err", err)
			return err
		}

		log.InfoContext(ctx, "deactivating Plaid link now")
		if err = client.RemoveItem(ctx); err != nil {
			log.ErrorContext(ctx, "failed to deactivate Plaid link", "err", err)
			return err
		}

		log.InfoContext(ctx, "Plaid link was successfully deactivated")
		return nil
	})
}
