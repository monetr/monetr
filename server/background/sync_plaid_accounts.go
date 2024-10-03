package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
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

func (s *SyncPlaidAccountsJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()
	crumbs.AddTag(span.Context(), "accountId", s.args.AccountId.String())
	crumbs.AddTag(span.Context(), "linkId", s.args.LinkId.String())

	log := s.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"accountId": s.args.AccountId,
		"linkId":    s.args.LinkId,
	})

	link, err := s.repo.GetLink(span.Context(), s.args.LinkId)
	if err = errors.Wrap(err, "failed to retrieve link to sync with plaid"); err != nil {
		log.WithError(err).Error("cannot sync without link")
		return err
	}

	if link.PlaidLink == nil {
		log.Warn("provided link does not have any plaid credentials")
		crumbs.IndicateBug(
			span.Context(),
			"BUG: Link was queued to sync with plaid, but has no plaid details",
			map[string]interface{}{
				"link": link,
			},
		)
		span.Status = sentry.SpanStatusFailedPrecondition
		return nil
	}

	log = log.WithFields(logrus.Fields{
		"plaidLinkId": link.PlaidLink.PlaidLinkId,
		"plaid": logrus.Fields{
			"institutionId":   link.PlaidLink.InstitutionId,
			"institutionName": link.PlaidLink.InstitutionName,
			"itemId":          link.PlaidLink.PlaidId,
		},
	})

	// This way other methods will have these log fields too.
	s.log = log

	plaidLink := link.PlaidLink

	bankAccounts, err := s.repo.GetBankAccountsWithPlaidByLinkId(
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

	if len(bankAccounts) == 0 {
		log.Warn("no bank accounts for plaid link")
		crumbs.Debug(span.Context(), "No bank accounts setup for plaid link", nil)
		return nil
	}

	secret, err := s.secrets.Read(span.Context(), plaidLink.SecretId)
	if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
		log.WithError(err).Error("could not retrieve API credentials for Plaid for link, this job will be retried")
		return err
	}

	plaidClient, err := s.plaidPlatypus.NewClient(
		span.Context(),
		link,
		secret.Value,
		plaidLink.PlaidId,
	)
	if err != nil {
		log.WithError(err).Error("failed to create plaid client for link")
		return err
	}

	plaidBankAccounts, err := plaidClient.GetAccounts(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to retrieve bank accounts from plaid")
		return err
	}

	{ // Handle missing or inactive accounts
		missingAccounts := map[ID[BankAccount]]*BankAccount{}
	MissingAccounts:
		for x := range bankAccounts {
			bankAccount := bankAccounts[x]
			for y := range plaidBankAccounts {
				plaidBankAccount := plaidBankAccounts[y]
				if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
					log.WithFields(logrus.Fields{
						"bankAccountId": bankAccount.BankAccountId,
					}).Debug("bank account is still present in plaid and is considered active")
					// TODO Check bank account status here too, if the status is inactive
					// but we see the account again then that means it is active again.
					continue MissingAccounts
				}
			}

			if bankAccount.Status == InactiveBankAccountStatus {
				// Bank account is already considered missing, skip it.
				continue
			}

			log.WithFields(logrus.Fields{
				"bankAccountId":       bankAccount.BankAccountId,
				"plaid_bankAccountId": bankAccount.PlaidBankAccount.PlaidId,
			}).Info("bank account is no longer present in plaid and is considered inactive")
			missingAccounts[bankAccount.BankAccountId] = &bankAccount
		}
	}

	return nil
}
