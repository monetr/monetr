package background

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	PullBalances = "PullBalances"
)

var (
	_ ScheduledJobHandler = &PullBalancesHandler{}
	_ Job                 = &PullBalancesJob{}
)

type (
	PullBalancesHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		jobRepo       repository.JobRepository
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
		unmarshaller  JobUnmarshaller
	}

	PullBalancesArguments struct {
		AccountId           uint64   `json:"accountId"`
		LinkId              uint64   `json:"linkId"`
		PlaidBankAccountIds []string `json:"plaidBankAccountIds"`
	}

	PullBalancesJob struct {
		args          PullBalancesArguments
		log           *logrus.Entry
		repo          repository.BaseRepository
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
	}
)

func TriggerPullBalances(ctx context.Context, backgroundJobs JobController, arguments PullBalancesArguments) error {
	return backgroundJobs.triggerJob(ctx, PullBalances, arguments)
}

func NewPullBalancesHandler(
	log *logrus.Entry,
	db *pg.DB,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
) *PullBalancesHandler {
	return &PullBalancesHandler{
		log:           log,
		db:            db,
		jobRepo:       repository.NewJobRepository(db),
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
		unmarshaller:  DefaultJobUnmarshaller,
	}
}

func (p *PullBalancesHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := p.log.WithContext(ctx)

	log.Info("retrieving links to pull balances for")

	accounts, err := p.jobRepo.GetPlaidLinksByAccount(ctx)
	if err != nil {
		log.WithError(err).Error("failed to retrieve bank accounts that need to be synced")
		return err
	}

	log.Infof("enqueueing %d account(s) for sync", len(accounts))

	for _, account := range accounts {
		for _, linkId := range account.LinkIds {
			accountLog := log.WithFields(logrus.Fields{
				"accountId": account.AccountId,
				"linkId":    linkId,
			})
			accountLog.Trace("enqueueing for account balance update")

			err = enqueuer.EnqueueJob(ctx, p.QueueName(), PullBalancesArguments{
				AccountId:           account.AccountId,
				LinkId:              linkId,
				PlaidBankAccountIds: nil, // All bank accounts.
			})
			if err != nil {
				log.WithError(err).Warn("failed to enqueue job to sync bank account balances")
				crumbs.Warn(ctx, "Failed to enqueue pull bank account balances job", "job", map[string]interface{}{
					"error": err,
				})
				continue
			}

			accountLog.Trace("successfully enqueued account for balance syncing")
		}
	}

	return nil
}

func (p PullBalancesHandler) DefaultSchedule() string {
	return "0 0 0 * * *" // Once a day at midnight in UTC.
}

func (p PullBalancesHandler) QueueName() string {
	return PullBalances
}

func (p PullBalancesHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args PullBalancesArguments
	if err := errors.Wrap(p.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Pull Balances job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       strconv.FormatUint(args.AccountId, 10),
				Username: fmt.Sprintf("account:%d", args.AccountId),
			})
		})
	}

	return p.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		repo := repository.NewRepositoryFromSession(0, args.AccountId, txn)
		job, err := NewPullBalancesJob(p.log.WithContext(ctx), repo, p.plaidSecrets, p.plaidPlatypus, args)
		if err != nil {
			return err
		}
		return job.Run(span.Context())
	})
}

func NewPullBalancesJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
	args PullBalancesArguments,
) (*PullBalancesJob, error) {
	return &PullBalancesJob{
		args:          args,
		log:           log,
		repo:          repo,
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
	}, nil
}

func (p *PullBalancesJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	start := time.Now()

	log := p.log.WithContext(span.Context())

	link, err := p.repo.GetLink(span.Context(), p.args.LinkId)
	if err != nil {
		log.WithError(err).Error("failed to retrieve link details to pull balances")
		return err
	}

	log = log.WithField("linkId", link.LinkId)

	if link.PlaidLink == nil {
		err = errors.Errorf("cannot pull account balanaces for link without plaid info")
		log.WithError(err).Errorf("failed to pull balances")
		return nil
	}

	switch link.LinkStatus {
	case models.LinkStatusSetup, models.LinkStatusPendingExpiration:
		break
	default:
		crumbs.Warn(span.Context(), "Link is not in a state where data can be retrieved", "plaid", map[string]interface{}{
			"status": link.LinkStatus,
		})
		return nil
	}

	accessToken, err := p.plaidSecrets.GetAccessTokenForPlaidLinkId(span.Context(), p.args.AccountId, link.PlaidLink.ItemId)
	if err != nil || accessToken == "" {
		log.WithError(err).Errorf("failed to retrieve access token for link")
		crumbs.Error(span.Context(), "Could not retrieve Plaid access token for link", "plaid", nil)
		return nil
	}

	bankAccounts, err := p.repo.GetBankAccountsByLinkId(span.Context(), p.args.LinkId)
	if err != nil {
		log.WithError(err).Error("failed to retrieve bank account details to pull balances")
		return err
	}

	// Gather the plaid account Ids so we can precisely query plaid.
	plaidIdsToBank := map[string]models.BankAccount{}
	itemBankAccountIds := make([]string, len(bankAccounts))
	for i, bankAccount := range bankAccounts {
		itemBankAccountIds[i] = bankAccount.PlaidAccountId
		plaidIdsToBank[bankAccount.PlaidAccountId] = bankAccount
	}

	log.Debugf("requesting information for %d bank account(s)", len(itemBankAccountIds))

	client, err := p.plaidPlatypus.NewClient(span.Context(), link, accessToken, link.PlaidLink.ItemId)
	if err != nil {
		log.WithError(err).Error("failed to create plaid client")
		return err
	}

	result, err := client.GetAccounts(
		span.Context(),
		itemBankAccountIds...,
	)
	if err != nil {
		log.WithError(err).Error("failed to retrieve bank accounts from plaid")
		return errors.Wrap(err, "failed to retrieve bank accounts from plaid")
	}

	updatedBankAccounts := make([]models.BankAccount, 0, len(result))
	for _, item := range result {
		bankAccount := plaidIdsToBank[item.GetAccountId()]
		bankLog := log.WithFields(logrus.Fields{
			"bankAccountId": bankAccount.BankAccountId,
			"linkId":        bankAccount.LinkId,
		})
		shouldUpdate := false
		available := item.GetBalances().GetAvailable()
		current := item.GetBalances().GetCurrent()

		if bankAccount.CurrentBalance != current {
			bankLog = bankLog.WithField("currentBalanceChanged", true)
			shouldUpdate = true
		} else {
			bankLog = bankLog.WithField("currentBalanceChanged", false)
		}

		if bankAccount.AvailableBalance != available {
			bankLog = bankLog.WithField("availableBalanceChanged", true)
			shouldUpdate = true
		} else {
			bankLog = bankLog.WithField("availableBalanceChanged", false)
		}

		bankLog = bankLog.WithField("willUpdate", shouldUpdate)

		if shouldUpdate {
			bankLog.Info("updating bank account balances")
		} else {
			bankLog.Trace("balances do not need to be updated")
		}

		if shouldUpdate {
			updatedBankAccounts = append(updatedBankAccounts, models.BankAccount{
				BankAccountId:    bankAccount.BankAccountId,
				AccountId:        p.args.AccountId,
				AvailableBalance: available,
				CurrentBalance:   current,
				LastUpdated:      start.UTC(),
			})
		}
	}

	if err = p.repo.UpdateBankAccounts(span.Context(), updatedBankAccounts); err != nil {
		log.WithError(err).Error("failed to update bank account balances")
		return err
	}

	return nil
}
