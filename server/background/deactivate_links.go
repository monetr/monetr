package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	DeactivateLinks = "DeactivateLinks"
)

type (
	DeactivateLinksHandler struct {
		log           *logrus.Entry
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
		log           *logrus.Entry
		repo          repository.BaseRepository
		secrets       repository.SecretsRepository
		plaidPlatypus platypus.Platypus
		clock         clock.Clock
	}
)

func NewDeactivateLinksHandler(
	log *logrus.Entry,
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
	log *logrus.Entry,
	data []byte,
) error {
	var args DeactivateLinksArguments
	if err := errors.Wrap(d.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Deactivate Links job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return d.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context())
		repo := repository.NewRepositoryFromSession(d.clock, "user_system", args.AccountId, txn)
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
	log := d.log.WithContext(ctx)

	if !d.configuration.Stripe.IsBillingEnabled() {
		log.Debug("billing is not enabled, plaid links will not automatically be deactivated")
		crumbs.Debug(ctx, "Billing is not enabled, plaid links will not automatically be deactivated", nil)
		return nil
	}

	log.Info("retrieving links for expired accounts")
	expiredLinks, err := d.repo.GetLinksForExpiredAccounts(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve funding schedules to process")
	}

	if len(expiredLinks) == 0 {
		crumbs.Debug(ctx, "No expired links were found at this time, no work to be done.", nil)
		log.Info("no expired links were found at this time, no work to be done")
		return nil
	}

	log.WithField("count", len(expiredLinks)).Info("preparing to enqueue jobs to remove expired links")

	for _, item := range expiredLinks {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
			"linkId":    item.LinkId,
		})
		itemLog.Trace("enqueuing job to remove expired link for account")
		err = enqueuer.EnqueueJob(ctx, d.QueueName(), DeactivateLinksArguments{
			AccountId: item.AccountId,
			LinkId:    item.LinkId,
		})
		if err != nil {
			log.WithError(err).Warn("failed to enqueue job to remove expired link")
			crumbs.Warn(ctx, "Failed to enqueue job to remove expired link", "job", map[string]interface{}{
				"error": err,
			})
			continue
		}

		itemLog.Trace("successfully enqueued removal of expired link")
	}

	return nil
}

func NewDeactivateLinksJob(
	log *logrus.Entry,
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

	log := d.log.WithContext(span.Context())

	link, err := d.repo.GetLink(span.Context(), d.args.LinkId)
	if err != nil {
		log.WithError(err).Error("failed to retrieve link for deactivation")
		return err
	}

	crumbs.IncludeUserInScope(span.Context(), link.AccountId)

	if link.PlaidLink == nil {
		log.Warn("provided link does not have any plaid credentials")
		crumbs.Warn(span.Context(), "BUG: Link was queued to be deactivated, but has no plaid details", "jobs", map[string]interface{}{
			"link": link,
		})
		span.Status = sentry.SpanStatusFailedPrecondition
		return nil
	}

	crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.PlaidId)

	secret, err := d.secrets.Read(span.Context(), link.PlaidLink.SecretId)
	if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
		log.WithError(err).Error("could not retrieve API credentials for Plaid for link, this job will be retried")
		return err
	}

	client, err := d.plaidPlatypus.NewClient(span.Context(), link, secret.Value, link.PlaidLink.PlaidId)
	if err != nil {
		log.WithError(err).Error("failed to create client for link deactivation")
		return err
	}

	if err = d.repo.DeletePlaidLink(span.Context(), *link.PlaidLinkId); err != nil {
		log.WithError(err).Warn("failed to remove Plaid details, link cannot be removed at this time")
		return err
	}

	log.Info("deactivating Plaid link now")
	if err = client.RemoveItem(span.Context()); err != nil {
		log.WithError(err).Error("failed to deactivate Plaid link, the job will not be retried")
		return nil
	}

	log.Info("Plaid link was successfully deactivated, removing Plaid details now")

	if err = d.secrets.Delete(span.Context(), secret.SecretId); err != nil {
		log.WithError(err).Error("failed to remove Plaid credentials for link")
		return nil // Don't retry.
	}

	return nil
}
