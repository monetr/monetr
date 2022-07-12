package background

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
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
		repo          repository.JobRepository
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
		unmarshaller  JobUnmarshaller
	}

	DeactivateLinksArguments struct {
		AccountId uint64 `json:"accountId"`
		LinkId    uint64 `json:"linkId"`
	}

	DeactivateLinksJob struct {
		args          DeactivateLinksArguments
		log           *logrus.Entry
		repo          repository.BaseRepository
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
	}
)

func NewDeactivateLinksHandler(
	log *logrus.Entry,
	db *pg.DB,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
) *DeactivateLinksHandler {
	return &DeactivateLinksHandler{
		log:           log,
		db:            db,
		repo:          repository.NewJobRepository(db),
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
		unmarshaller:  DefaultJobUnmarshaller,
	}
}

func (d *DeactivateLinksHandler) SetUnmarshaller(unmarshaller JobUnmarshaller) {
	d.unmarshaller = unmarshaller
}

func (d DeactivateLinksHandler) QueueName() string {
	return DeactivateLinks
}

func (d *DeactivateLinksHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
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

		repo := repository.NewRepositoryFromSession(0, args.AccountId, txn)
		job, err := NewDeactivateLinksJob(
			d.log.WithContext(span.Context()),
			repo,
			d.plaidSecrets,
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

	jobErrors := make([]error, 0)

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
			jobErrors = append(jobErrors, err)
			continue
		}

		itemLog.Trace("successfully enqueued removal of expired link")
	}

	return nil
}

func NewDeactivateLinksJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
	args DeactivateLinksArguments,
) (*DeactivateLinksJob, error) {
	return &DeactivateLinksJob{
		args:          args,
		log:           log,
		repo:          repo,
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
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

	if link.PlaidLink == nil {
		log.Warn("provided link does not have any plaid credentials")
		crumbs.Warn(span.Context(), "BUG: Link was queued to be deactivated, but has no plaid details", "jobs", map[string]interface{}{
			"link": link,
		})
		span.Status = sentry.SpanStatusFailedPrecondition
		return nil
	}

	crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.ItemId)

	accessToken, err := d.plaidSecrets.GetAccessTokenForPlaidLinkId(span.Context(), d.args.AccountId, link.PlaidLink.ItemId)
	if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
		// If the token is simply missing from vault then something is goofy. Don't retry the job but mark it as a
		// failure.
		if errors.Is(errors.Cause(err), secrets.ErrNotFound) {
			if hub := sentry.GetHubFromContext(span.Context()); hub != nil {
				hub.ConfigureScope(func(scope *sentry.Scope) {
					// Mark the scope as an error.
					scope.SetLevel(sentry.LevelError)
				})
			}

			log.WithError(err).Error("could not retrieve API credentials for Plaid for link, job will not be retried")
			return nil
		}

		log.WithError(err).Error("could not retrieve API credentials for Plaid for link, this job will be retried")
		return err
	}

	client, err := d.plaidPlatypus.NewClient(span.Context(), link, accessToken, link.PlaidLink.ItemId)
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

	if err = d.plaidSecrets.RemoveAccessTokenForPlaidLink(span.Context(), d.args.AccountId, link.PlaidLink.ItemId); err != nil {
		log.WithError(err).Error("failed to remove Plaid credentials for link")
		return nil // Don't retry.
	}

	return nil
}
