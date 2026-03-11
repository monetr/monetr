package background

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
)

const (
	DeactivateLinks = "DeactivateLinks"
)

type (
	DeactivateLinksHandler struct {
		log           *slog.Logger
		db            *pg.DB
		configuration config.Configuration
		repo          repository.JobRepository
		kms           secrets.KeyManagement
		plaidPlatypus platypus.Platypus
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	DeactivateLinksArguments struct {
		AccountId ID[Account] `json:"accountId"`
		LinkId    ID[Link]    `json:"linkId"`
	}

	DeactivateLinksJob struct {
		args          DeactivateLinksArguments
		log           *slog.Logger
		repo          repository.BaseRepository
		secrets       repository.SecretsRepository
		plaidPlatypus platypus.Platypus
		clock         clock.Clock
	}
)

func NewDeactivateLinksHandler(
	log *slog.Logger,
	db *pg.DB,
	clock clock.Clock,
	configuration config.Configuration,
	kms secrets.KeyManagement,
	plaidPlatypus platypus.Platypus,
) *DeactivateLinksHandler {
	return &DeactivateLinksHandler{
		log:           log,
		db:            db,
		configuration: configuration,
		repo:          repository.NewJobRepository(db, clock),
		kms:           kms,
		plaidPlatypus: plaidPlatypus,
		unmarshaller:  DefaultJobUnmarshaller,
		clock:         clock,
	}
}

func (d DeactivateLinksHandler) QueueName() string {
	return DeactivateLinks
}

func (d *DeactivateLinksHandler) HandleConsumeJob(
	ctx context.Context,
	log *slog.Logger,
	data []byte,
) error {
	var args DeactivateLinksArguments
	if err := errors.Wrap(d.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Deactivate Links job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return d.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log = log.With(
			"accountId", args.AccountId,
			"linkId", args.LinkId,
		)
		repo := repository.NewRepositoryFromSession(
			d.clock,
			"user_system",
			args.AccountId,
			txn,
			log,
		)
		secretsRepo := repository.NewSecretsRepository(
			log,
			d.clock,
			txn,
			d.kms,
			args.AccountId,
		)
		job, err := NewDeactivateLinksJob(
			log,
			repo,
			d.clock,
			secretsRepo,
			d.plaidPlatypus,
			args,
		)
		if err != nil {
			return err
		}
		return job.Run(span.Context())
	})
}

func (d DeactivateLinksHandler) DefaultSchedule() string {
	// Will run once a day.
	return "0 0 0 * * *"
}

func (d *DeactivateLinksHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := d.log

	if !d.configuration.Stripe.IsBillingEnabled() {
		log.DebugContext(ctx, "billing is not enabled, plaid links will not automatically be deactivated")
		crumbs.Debug(ctx, "Billing is not enabled, plaid links will not automatically be deactivated", nil)
		return nil
	}

	log.InfoContext(ctx, "retrieving links for expired accounts")
	expiredLinks, err := d.repo.GetLinksForExpiredAccounts(ctx)
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
		err = enqueuer.EnqueueJob(ctx, d.QueueName(), DeactivateLinksArguments{
			AccountId: item.AccountId,
			LinkId:    item.LinkId,
		})
		if err != nil {
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

func NewDeactivateLinksJob(
	log *slog.Logger,
	repo repository.BaseRepository,
	clock clock.Clock,
	secrets repository.SecretsRepository,
	plaidPlatypus platypus.Platypus,
	args DeactivateLinksArguments,
) (*DeactivateLinksJob, error) {
	return &DeactivateLinksJob{
		args:          args,
		log:           log,
		repo:          repo,
		secrets:       secrets,
		plaidPlatypus: plaidPlatypus,
		clock:         clock,
	}, nil
}

func (d *DeactivateLinksJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := d.log

	link, err := d.repo.GetLink(span.Context(), d.args.LinkId)
	if err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve link for deactivation", "err", err)
		return err
	}

	crumbs.IncludeUserInScope(span.Context(), link.AccountId)

	if link.PlaidLink == nil {
		log.WarnContext(span.Context(), "provided link does not have any plaid credentials")
		crumbs.Warn(span.Context(), "BUG: Link was queued to be deactivated, but has no plaid details", "jobs", map[string]any{
			"link": link,
		})
		span.Status = sentry.SpanStatusFailedPrecondition
		return nil
	}

	crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.PlaidId)

	log = log.With(
		"accountId", d.args.AccountId,
		"linkId", d.args.LinkId,
		"plaidLinkId", link.PlaidLinkId,
		"itemId", link.PlaidLink.PlaidId,
		"institutionId", link.PlaidLink.InstitutionId,
		"institutionName", link.PlaidLink.InstitutionName,
		"plaidLinkStatus", link.PlaidLink.Status.String(),
	)

	log.InfoContext(span.Context(), "removing plaid link")

	secret, err := d.secrets.Read(span.Context(), link.PlaidLink.SecretId)
	if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
		log.ErrorContext(span.Context(), "could not retrieve API credentials for Plaid for link, this job will be retried", "err", err)
		return err
	}

	client, err := d.plaidPlatypus.NewClient(span.Context(), link, secret.Value, link.PlaidLink.PlaidId)
	if err != nil {
		log.ErrorContext(span.Context(), "failed to create client for link deactivation", "err", err)
		return err
	}

	if err = d.repo.DeletePlaidLink(span.Context(), *link.PlaidLinkId); err != nil {
		log.WarnContext(span.Context(), "failed to remove Plaid details, link cannot be removed at this time", "err", err)
		return err
	}

	log.InfoContext(span.Context(), "deactivating Plaid link now")
	if err = client.RemoveItem(span.Context()); err != nil {
		log.ErrorContext(span.Context(), "failed to deactivate Plaid link, the job will not be retried", "err", err)
		return nil
	}

	log.InfoContext(span.Context(), "Plaid link was successfully deactivated")

	return nil
}

func DeactivateLinksCron(ctx queue.Context) error {
	log := ctx.Log()

	log.InfoContext(ctx, "retrieving links for expired accounts")
	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())
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
		if err = queue.Enqueue(ctx, ctx.Processor(), DeactivateLink, DeactivateLinksArguments{
			AccountId: item.AccountId,
			LinkId:    item.LinkId,
		}); err != nil {
			itemLog.WarnContext(ctx, "failed to enqueue job to remove expired link", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to remove expired link", "job", map[string]any{
				"error": err,
			})
			continue
		}
		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued removal of expired link")
	}

	return nil
}

func DeactivateLink(ctx queue.Context, args DeactivateLinksArguments) error {
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
			"plaidLinkStatus", link.PlaidLink.Status.String(),
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
			log.ErrorContext(ctx, "failed to deactivate Plaid link, the job will not be retried", "err", err)
			return nil
		}

		log.InfoContext(ctx, "Plaid link was successfully deactivated")
		return nil
	})
}
