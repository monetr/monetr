package background

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	PullTransactions = "PullTransactions"
)

var (
	_ JobHandler = &PullTransactionsHandler{}
	_ Job        = &PullTransactionsJob{}
)

type (
	PullTransactionsHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
		publisher     pubsub.Publisher
		unmarshaller  JobUnmarshaller
	}

	PullTransactionsArguments struct {
		AccountId uint64    `json:"accountId"`
		LinkId    uint64    `json:"linkId"`
		Start     time.Time `json:"start"`
		End       time.Time `json:"end"`
	}

	PullTransactionsJob struct {
		args          PullTransactionsArguments
		log           *logrus.Entry
		repo          repository.BaseRepository
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
		publisher     pubsub.Publisher
	}
)

func TriggerPullTransactions(ctx context.Context, backgroundJobs JobController, arguments PullTransactionsArguments) error {
	return backgroundJobs.triggerJob(ctx, PullTransactions, arguments)
}

func NewPullTransactionsHandler(
	log *logrus.Entry,
	db *pg.DB,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
	publisher pubsub.Publisher,
) *PullTransactionsHandler {
	return &PullTransactionsHandler{
		log:           log,
		db:            db,
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
		publisher:     publisher,
		unmarshaller:  DefaultJobUnmarshaller,
	}
}

func (p PullTransactionsHandler) QueueName() string {
	return PullTransactions
}

func (p *PullTransactionsHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args PullTransactionsArguments
	if err := errors.Wrap(p.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Pull Transactions job.", "job", map[string]interface{}{
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
		job, err := NewPullTransactionsJob(
			p.log.WithContext(span.Context()),
			repo,
			p.plaidSecrets,
			p.plaidPlatypus,
			p.publisher,
			args,
		)
		if err != nil {
			return err
		}
		return job.Run(span.Context())
	})
}

func NewPullTransactionsJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
	publisher pubsub.Publisher,
	args PullTransactionsArguments,
) (*PullTransactionsJob, error) {
	return &PullTransactionsJob{
		args:          args,
		log:           log,
		repo:          repo,
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
		publisher:     publisher,
	}, nil
}

func (p *PullTransactionsJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := p.log.WithContext(span.Context())

	link, err := p.repo.GetLink(span.Context(), p.args.LinkId)
	if err = errors.Wrap(err, "failed to retrieve link to pull transactions"); err != nil {
		log.WithError(err).Error("cannot retrieve transactions without link")
		return err
	}

	if link.PlaidLink == nil {
		log.Warn("provided link does not have any plaid credentials")
		crumbs.Warn(span.Context(), "BUG: Link was queued to pull transactions, but has no plaid details", "jobs", map[string]interface{}{
			"link": link,
		})
		span.Status = sentry.SpanStatusFailedPrecondition
		return nil
	}

	crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.ItemId)

	if len(link.BankAccounts) == 0 {
		log.Warn("no bank accounts for plaid link")
		crumbs.Debug(span.Context(), "No bank accounts setup for plaid link", nil)
		return nil
	}

	accessToken, err := p.plaidSecrets.GetAccessTokenForPlaidLinkId(span.Context(), p.args.AccountId, link.PlaidLink.ItemId)
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

	plaidIdsToBankIds := map[string]uint64{}
	plaidBankToLocalBank := map[string]models.BankAccount{}
	bankAccountIds := make([]string, len(link.BankAccounts))
	for i, bankAccount := range link.BankAccounts {
		bankAccountIds[i] = bankAccount.PlaidAccountId
		plaidIdsToBankIds[bankAccount.PlaidAccountId] = bankAccount.BankAccountId
		plaidBankToLocalBank[bankAccount.PlaidAccountId] = bankAccount
	}

	now := time.Now().UTC()
	plaidClient, err := p.plaidPlatypus.NewClient(span.Context(), link, accessToken, link.PlaidLink.ItemId)
	if err != nil {
		log.WithError(err).Error("failed to create plaid client for link")
		return err
	}

	plaidTransactions, err := plaidClient.GetAllTransactions(span.Context(), p.args.Start, p.args.End, bankAccountIds)
	if err = errors.Wrap(err, "failed to retrieve transactions from plaid for sync"); err != nil {
		log.WithError(err).Error("failed to pull transactions")
		return err
	}

	if len(plaidTransactions) == 0 {
		log.Warn("no transactions were retrieved from plaid")
		crumbs.Debug(span.Context(), "No transactions were retrieved from plaid.", nil)
	}

	log.WithField("count", len(plaidTransactions)).Debugf("retrieved transactions from plaid")
	crumbs.Debug(span.Context(), "Retrieved transactions from plaid.", map[string]interface{}{
		"count": len(plaidTransactions),
	})

	account, err := p.repo.GetAccount(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to retrieve account for job")
		return err
	}

	timezone, err := account.GetTimezone()
	if err != nil {
		log.WithError(err).Warn("failed to get account's time zone, defaulting to UTC")
		timezone = time.UTC
	}

	plaidTransactionIds := make([]string, len(plaidTransactions))
	for i, transaction := range plaidTransactions {
		plaidTransactionIds[i] = transaction.GetTransactionId()
	}

	transactionsByPlaidId, err := p.repo.GetTransactionsByPlaidId(span.Context(), link.LinkId, plaidTransactionIds)
	if err != nil {
		log.WithError(err).Error("failed to retrieve transaction ids for updating plaid transactions")
		return err
	}

	log.Debugf("found %d existing transactions", len(transactionsByPlaidId))

	transactionsToUpdate := make([]*models.Transaction, 0)
	transactionsToInsert := make([]models.Transaction, 0)
	for _, plaidTransaction := range plaidTransactions {
		amount := plaidTransaction.GetAmount()

		date := plaidTransaction.GetDateLocal(timezone)

		transactionName := plaidTransaction.GetName()

		// We only want to make the transaction name be the merchant name if the merchant name is shorter. This is
		// due to something I observed with a dominos transaction, where the merchant was improperly parsed and the
		// transaction ended up being called `Mnuslindstrom` rather than `Domino's`. This should fix that problem.
		if plaidTransaction.GetMerchantName() != "" && len(plaidTransaction.GetMerchantName()) < len(transactionName) {
			transactionName = plaidTransaction.GetMerchantName()
		}

		existingTransaction, ok := transactionsByPlaidId[plaidTransaction.GetTransactionId()]
		if !ok {
			transactionsToInsert = append(transactionsToInsert, models.Transaction{
				AccountId:                 p.repo.AccountId(),
				BankAccountId:             plaidIdsToBankIds[plaidTransaction.GetBankAccountId()],
				PlaidTransactionId:        plaidTransaction.GetTransactionId(),
				Amount:                    amount,
				SpendingId:                nil,
				Spending:                  nil,
				Categories:                plaidTransaction.GetCategory(),
				OriginalCategories:        plaidTransaction.GetCategory(),
				Date:                      date,
				Name:                      transactionName,
				OriginalName:              plaidTransaction.GetName(),
				MerchantName:              plaidTransaction.GetMerchantName(),
				OriginalMerchantName:      plaidTransaction.GetMerchantName(),
				IsPending:                 plaidTransaction.GetIsPending(),
				CreatedAt:                 now,
				PendingPlaidTransactionId: plaidTransaction.GetPendingTransactionId(),
			})
			continue
		}

		var shouldUpdate bool
		if existingTransaction.Amount != amount {
			shouldUpdate = true
		}

		if existingTransaction.IsPending != plaidTransaction.GetIsPending() {
			shouldUpdate = true
		}

		if !myownsanity.StringPEqual(existingTransaction.PendingPlaidTransactionId, plaidTransaction.GetPendingTransactionId()) {
			shouldUpdate = true
		}

		existingTransaction.Amount = amount
		existingTransaction.IsPending = plaidTransaction.GetIsPending()
		existingTransaction.PendingPlaidTransactionId = plaidTransaction.GetPendingTransactionId()

		// Fix timezone of records.
		if !existingTransaction.Date.Equal(date) {
			existingTransaction.Date = date
			shouldUpdate = true
		}

		if shouldUpdate {
			transactionsToUpdate = append(transactionsToUpdate, &existingTransaction)
		}
	}

	if len(transactionsToUpdate) > 0 {
		log.Infof("updating %d transactions", len(transactionsToUpdate))
		crumbs.Debug(span.Context(), "Updating transactions.", map[string]interface{}{
			"count": len(transactionsToUpdate),
		})
		if err = p.repo.UpdateTransactions(span.Context(), transactionsToUpdate); err != nil {
			log.WithError(err).Errorf("failed to update transactions for job")
			return err
		}
	}

	if len(transactionsToInsert) > 0 {
		log.Infof("creating %d transactions", len(transactionsToInsert))
		crumbs.Debug(span.Context(), "Creating transactions.", map[string]interface{}{
			"count": len(transactionsToInsert),
		})
		// Reverse the list so the oldest records are inserted first.
		for i, j := 0, len(transactionsToInsert)-1; i < j; i, j = i+1, j-1 {
			transactionsToInsert[i], transactionsToInsert[j] = transactionsToInsert[j], transactionsToInsert[i]
		}
		if err = p.repo.InsertTransactions(span.Context(), transactionsToInsert); err != nil {
			log.WithError(err).Error("failed to insert new transactions")
			return err
		}
	}

	if len(transactionsToInsert)+len(transactionsToUpdate) > 0 {
		result, err := plaidClient.GetAccounts(
			span.Context(),
			bankAccountIds...,
		)
		if err != nil {
			log.WithError(err).Error("failed to retrieve bank accounts from plaid")
			return errors.Wrap(err, "failed to retrieve bank accounts from plaid")
		}

		updatedBankAccounts := make([]models.BankAccount, 0, len(result))
		for _, item := range result {
			bankAccount, ok := plaidBankToLocalBank[item.GetAccountId()]
			if !ok {
				log.WithField("plaidBankAccountId", item.GetAccountId()).Warn("bank was not found in map")
				continue
			}

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
					LastUpdated:      now.UTC(),
				})
			}
		}

		if err = p.repo.UpdateBankAccounts(span.Context(), updatedBankAccounts); err != nil {
			log.WithError(err).Error("failed to update bank account balances")
			crumbs.ReportError(span.Context(), err, "Failed to update bank account balances", "job", nil)
		}
	}

	linkWasSetup := false

	// If the link status is not setup or pending expiration. Then change the status to setup
	switch link.LinkStatus {
	case models.LinkStatusSetup, models.LinkStatusPendingExpiration:
	default:
		crumbs.Debug(span.Context(), "Updating link status.", map[string]interface{}{
			"old": link.LinkStatus,
			"new": models.LinkStatusSetup,
		})
		link.LinkStatus = models.LinkStatusSetup
		linkWasSetup = true
	}
	link.LastSuccessfulUpdate = myownsanity.TimeP(time.Now().UTC())
	if err = p.repo.UpdateLink(span.Context(), link); err != nil {
		log.WithError(err).Error("failed to update link after transaction sync")
		return err
	}

	if linkWasSetup { // Send the notification that the link has been set up.
		channelName := fmt.Sprintf("initial:plaid:link:%d:%d", p.args.AccountId, p.args.LinkId)
		if notifyErr := p.publisher.Notify(
			span.Context(),
			channelName,
			"success",
		); notifyErr != nil {
			log.WithError(notifyErr).Error("failed to publish link status to pubsub")
		}
	}

	return nil
}
