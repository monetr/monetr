package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	SyncPlaidAccounts = "SyncPlaidAccounts"
)

type (
	SyncPlaidAccountsHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		kms           secrets.KeyManagement
		plaidPlatypus platypus.Platypus
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	SyncPlaidAccountsArguments struct {
		AccountId ID[Account] `json:"accountId"`
		LinkId    ID[Link]    `json:"linkId"`
	}

	SyncPlaidAccountsJob struct {
		args          SyncPlaidAccountsArguments
		log           *logrus.Entry
		repo          repository.BaseRepository
		secrets       repository.SecretsRepository
		plaidPlatypus platypus.Platypus
		clock         clock.Clock

		bankAccounts      []BankAccount
		plaidBankAccounts []platypus.BankAccount
	}
)

func TriggerSyncPlaidAccounts(
	ctx context.Context,
	backgroundJobs JobController,
	arguments SyncPlaidAccountsArguments,
) error {
	return backgroundJobs.EnqueueJob(ctx, SyncPlaidAccounts, arguments)
}

func NewSyncPlaidAccountsHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	kms secrets.KeyManagement,
	plaidPlatypus platypus.Platypus,
) *SyncPlaidAccountsHandler {
	return &SyncPlaidAccountsHandler{
		log:           log,
		db:            db,
		kms:           kms,
		plaidPlatypus: plaidPlatypus,
		unmarshaller:  DefaultJobUnmarshaller,
		clock:         clock,
	}
}

func (s SyncPlaidAccountsHandler) QueueName() string {
	return SyncPlaidAccounts
}

func (s *SyncPlaidAccountsHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args SyncPlaidAccountsArguments
	if err := errors.Wrap(s.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Sync Plaid accounts job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return s.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context()).WithFields(logrus.Fields{
			"accountId": args.AccountId,
			"linkId":    args.LinkId,
		})

		repo := repository.NewRepositoryFromSession(
			s.clock,
			"user_plaid",
			args.AccountId,
			txn,
			log,
		)
		secretsRepo := repository.NewSecretsRepository(
			log,
			s.clock,
			txn,
			s.kms,
			args.AccountId,
		)
		job, err := NewSyncPlaidAccountsJob(
			log,
			repo,
			s.clock,
			secretsRepo,
			s.plaidPlatypus,
			args,
		)
		if err != nil {
			return err
		}
		return job.Run(span.Context())
	})
}

func (s SyncPlaidAccountsHandler) DefaultSchedule() string {
	// Every 12 hours 15 minutes past the hour.
	return "0 15 */12 * * *"
}

func (s *SyncPlaidAccountsHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := s.log.WithContext(ctx)

	log.Info("retrieving links to sync with Plaid for updated accounts")

	links := make([]Link, 0)
	cutoff := s.clock.Now().AddDate(0, 0, -7)
	err := s.db.ModelContext(ctx, &links).
		Join(`INNER JOIN "plaid_links" AS "plaid_link"`).
		JoinOn(`"plaid_link"."plaid_link_id" = "link"."plaid_link_id"`).
		Where(`"plaid_link"."status" = ?`, PlaidLinkStatusSetup).
		Where(`"plaid_link"."last_account_sync" < ? OR "plaid_link"."last_account_sync" IS NULL`, cutoff).
		Where(`"plaid_link"."deleted_at" IS NULL`).
		Where(`"link"."link_type" = ?`, PlaidLinkType).
		Where(`"link"."deleted_at" IS NULL`).
		Select(&links)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve links that need to be synced with plaid for updated accounts")
	}

	if len(links) == 0 {
		log.Debug("no plaid links need to be synced at this time for updating accounts")
		return nil
	}

	log.WithField("count", len(links)).Info("syncing plaid links for accounts")

	for _, item := range links {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
			"linkId":    item.LinkId,
		})
		itemLog.Trace("enqueuing link to be synced with plaid for accounts")
		err := enqueuer.EnqueueJob(ctx, s.QueueName(), SyncPlaidAccountsArguments{
			AccountId: item.AccountId,
			LinkId:    item.LinkId,
		})
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job to sync with plaid accounts")
			crumbs.Warn(ctx, "Failed to enqueue job to sync with plaid accounts", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Trace("successfully enqueued link to be synced with plaid accounts")
	}

	return nil
}

func NewSyncPlaidAccountsJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	secrets repository.SecretsRepository,
	plaidPlatypus platypus.Platypus,
	args SyncPlaidAccountsArguments,
) (*SyncPlaidAccountsJob, error) {
	return &SyncPlaidAccountsJob{
		args:              args,
		log:               log,
		repo:              repo,
		secrets:           secrets,
		plaidPlatypus:     plaidPlatypus,
		clock:             clock,
		bankAccounts:      nil,
		plaidBankAccounts: nil,
	}, nil
}

func (j *SyncPlaidAccountsJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()
	crumbs.AddTag(span.Context(), "accountId", j.args.AccountId.String())
	crumbs.AddTag(span.Context(), "linkId", j.args.LinkId.String())

	log := j.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"accountId": j.args.AccountId,
		"linkId":    j.args.LinkId,
	})

	link, err := j.repo.GetLink(span.Context(), j.args.LinkId)
	if err = errors.Wrap(err, "failed to retrieve link to sync with plaid"); err != nil {
		log.WithError(err).Error("cannot sync without link")
		return err
	}

	if link.PlaidLink == nil {
		log.Warn("provided link does not have any plaid credentials")
		crumbs.IndicateBug(
			span.Context(),
			"BUG: Link was queued to sync with plaid, but has no plaid details",
			map[string]any{
				"link": link,
			},
		)
		span.Status = sentry.SpanStatusFailedPrecondition
		return nil
	}

	log = log.WithFields(logrus.Fields{
		"plaidLinkId":           link.PlaidLink.PlaidLinkId,
		"plaid_institutionId":   link.PlaidLink.InstitutionId,
		"plaid_institutionName": link.PlaidLink.InstitutionName,
		"plaid_itemId":          link.PlaidLink.PlaidId,
	})

	// This way other methods will have these log fields too.
	j.log = log

	plaidLink := link.PlaidLink

	j.bankAccounts, err = j.repo.GetBankAccountsWithPlaidByLinkId(
		span.Context(),
		link.LinkId,
	)
	if err = errors.Wrap(err, "failed to read bank accounts for plaid sync"); err != nil {
		log.WithError(err).Error("cannot sync without bank accounts")
		return err
	}
	crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.PlaidId)
	crumbs.AddTag(span.Context(), "plaid.institution_id", link.PlaidLink.InstitutionId)
	crumbs.AddTag(span.Context(), "plaid.institution_name", link.PlaidLink.InstitutionName)

	if len(j.bankAccounts) == 0 {
		log.Warn("no bank accounts for plaid link")
		crumbs.Debug(span.Context(), "No bank accounts setup for plaid link", nil)
		return nil
	}

	secret, err := j.secrets.Read(span.Context(), plaidLink.SecretId)
	if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
		log.WithError(err).Error("could not retrieve API credentials for Plaid for link, this job will be retried")
		return err
	}

	plaidClient, err := j.plaidPlatypus.NewClient(
		span.Context(),
		link,
		secret.Value,
		plaidLink.PlaidId,
	)
	if err != nil {
		log.WithError(err).Error("failed to create plaid client for link")
		return err
	}

	j.plaidBankAccounts, err = plaidClient.GetAccounts(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to retrieve bank accounts from plaid")
		return err
	}

	missingAccounts := j.findMissingAccounts()
	if len(missingAccounts) > 0 {
		log.WithFields(logrus.Fields{
			"count": len(missingAccounts),
		}).Info("found newly inactive accounts, updating status")

		for _, bankAccount := range missingAccounts {
			bankAccount.Status = InactiveBankAccountStatus
			j.log.WithFields(logrus.Fields{
				"bankAccountId":       bankAccount.BankAccountId,
				"plaid_bankAccountId": bankAccount.PlaidBankAccount.PlaidId,
			}).Debug("updating account to be inactive")
			if err := j.repo.UpdateBankAccount(span.Context(), &bankAccount); err != nil {
				log.WithError(err).Error("failed to mark account as inactive")
				continue
			}
		}
	} else {
		log.Info("no accounts to mark as inactive")
	}

	activeAccounts := j.findActiveAccounts()
	if len(activeAccounts) > 0 {
		log.WithFields(logrus.Fields{
			"count": len(activeAccounts),
		}).Info("found reactivated accounts, updating status")

		for _, bankAccount := range activeAccounts {
			bankAccount.Status = ActiveBankAccountStatus
			j.log.WithFields(logrus.Fields{
				"bankAccountId":       bankAccount.BankAccountId,
				"plaid_bankAccountId": bankAccount.PlaidBankAccount.PlaidId,
			}).Debug("updating account to be reactivated")
			if err := j.repo.UpdateBankAccount(span.Context(), &bankAccount); err != nil {
				log.WithError(err).Error("failed to mark account as active")
				continue
			}
		}
	} else {
		log.Info("no accounts to mark as reactivated")
	}

	log.Trace("updating plaid link's last account sync timestamp")
	plaidLink.LastAccountSync = myownsanity.TimeP(j.clock.Now())
	if err := j.repo.UpdatePlaidLink(span.Context(), plaidLink); err != nil {
		log.WithError(err).Error("failed to update plaid link's last account sync timestamp")
		return err
	}

	return nil
}

// findMissingAccounts will return a map of all the bank accounts who are
// currently in an active status, but who are no longer being returned by Plaid.
// These accounts will be marked as inactive.
func (j *SyncPlaidAccountsJob) findMissingAccounts() (missingAccounts map[ID[BankAccount]]BankAccount) {
	missingAccounts = make(map[ID[BankAccount]]BankAccount)
MissingAccounts:
	for x := range j.bankAccounts {
		bankAccount := j.bankAccounts[x]
		for y := range j.plaidBankAccounts {
			plaidBankAccount := j.plaidBankAccounts[y]
			if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
				j.log.WithFields(logrus.Fields{
					"bankAccountId":       bankAccount.BankAccountId,
					"plaid_bankAccountId": bankAccount.PlaidBankAccount.PlaidId,
				}).Debug("bank account is still present in plaid and is considered active")
				// TODO Check bank account status here too, if the status is inactive
				// but we see the account again then that means it is active again.
				continue MissingAccounts
			}
		}

		if bankAccount.Status == InactiveBankAccountStatus {
			// Bank account is already considered missing, skip it.
			j.log.WithFields(logrus.Fields{
				"bankAccountId":       bankAccount.BankAccountId,
				"plaid_bankAccountId": bankAccount.PlaidBankAccount.PlaidId,
			}).Trace("bank account is already inactive, it does not need to be updated")
			continue
		}

		j.log.WithFields(logrus.Fields{
			"bankAccountId":       bankAccount.BankAccountId,
			"plaid_bankAccountId": bankAccount.PlaidBankAccount.PlaidId,
		}).Info("bank account is no longer present in plaid and is considered inactive")
		missingAccounts[bankAccount.BankAccountId] = bankAccount
	}

	return missingAccounts
}

// findActiveAccounts will return a map of all the accounts who were previously
// marked as inactive but are now being seen in plaid's API responses again.
// This would be extremely unusual.
func (j *SyncPlaidAccountsJob) findActiveAccounts() (activeAcounts map[ID[BankAccount]]BankAccount) {
	activeAcounts = make(map[ID[BankAccount]]BankAccount)
ActiveAccounts:
	for x := range j.bankAccounts {
		bankAccount := j.bankAccounts[x]

		// If the account is already marked as active then skip it.
		if bankAccount.Status == ActiveBankAccountStatus {
			continue ActiveAccounts
		}

		for y := range j.plaidBankAccounts {
			plaidBankAccount := j.plaidBankAccounts[y]
			if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
				activeAcounts[bankAccount.BankAccountId] = bankAccount
				j.log.WithFields(logrus.Fields{
					"bankAccountId":       bankAccount.BankAccountId,
					"plaid_bankAccountId": bankAccount.PlaidBankAccount.PlaidId,
				}).Info("found inactive account that is present in Plaid again, will be updated to show as active")
				continue ActiveAccounts
			}
		}
	}

	return activeAcounts
}
