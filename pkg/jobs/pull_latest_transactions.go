package jobs

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	EnqueuePullLatestTransactions = "EnqueuePullLatestTransactions"
	PullLatestTransactions        = "PullLatestTransactions"
)

func (j *jobManagerBase) TriggerPullLatestTransactions(accountId, linkId uint64, numberOfTransactions int64) (jobId string, err error) {
	log := j.log.WithFields(logrus.Fields{
		"accountId": accountId,
		"linkId":    linkId,
	})

	log.Infof("queueing pull latest transactions for account")
	job, err := j.queue.EnqueueUnique(PullLatestTransactions, map[string]interface{}{
		"accountId":            accountId,
		"linkId":               linkId,
		"numberOfTransactions": numberOfTransactions,
	})
	if err != nil {
		log.WithError(err).Error("failed to enqueue pulling latest transactions")
		return "", errors.Wrap(err, "failed to enqueue pulling latest transactions")
	}
	log = log.WithField("pullLatestTransactionsJobId", job.ID)

	log.Infof("queueing account balances update for account")
	job, err = j.queue.EnqueueUnique(PullAccountBalances, map[string]interface{}{
		"accountId": accountId,
		"linkId":    linkId,
	})
	if err != nil {
		log.WithError(err).Error("failed to enqueue pulling account balances")
		return "", errors.Wrap(err, "failed to enqueue pulling account balances")
	}

	return job.ID, nil
}

func (j *jobManagerBase) enqueuePullLatestTransactions(job *work.Job) error {
	log := j.getLogForJob(job)

	accounts, err := j.getPlaidLinksByAccount()
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve bank accounts that need to by synced")
		return err
	}

	log.Infof("enqueing %d account(s) to pull latest transactions", len(accounts))

	for _, account := range accounts {
		for _, linkId := range account.LinkIDs {
			accountLog := log.WithFields(logrus.Fields{
				"accountId": account.AccountID,
				"linkId":    linkId,
			})
			accountLog.Trace("enqueueing for latest transactions update")
			_, err := j.queue.EnqueueUnique(PullLatestTransactions, map[string]interface{}{
				"accountId": account.AccountID,
				"linkId":    linkId,
			})
			if err != nil {
				accountLog.WithError(err).Error("could not enqueue account, data will not be synced")
				continue
			}

			accountLog.Trace("successfully enqueued account for latest transactions update")
		}
	}

	return nil
}

func (j *jobManagerBase) pullLatestTransactions(job *work.Job) error {
	span := sentry.StartSpan(context.Background(), "Job", sentry.TransactionName("Pull Latest Transactions"))
	defer span.Finish()

	start := time.Now()
	log := j.getLogForJob(job)
	log.Infof("pulling account balances")

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	defer func() {
		if j.stats != nil {
			j.stats.JobFinished(PullAccountBalances, accountId, start)
		}
	}()

	linkId := uint64(job.ArgInt64("linkId"))
	span.SetTag("linkId", strconv.FormatUint(linkId, 10))
	span.SetTag("accountId", strconv.FormatUint(accountId, 10))

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		account, err := repo.GetAccount()
		if err != nil {
			log.WithError(err).Error("failed to retrieve account for job")
			return err
		}

		timezone, err := account.GetTimezone()
		if err != nil {
			log.WithError(err).Warn("failed to get account's time zone, defaulting to UTC")
			timezone = time.UTC
		}

		link, err := repo.GetLink(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve link details to pull transactions")
			return err
		}

		if link.PlaidLink == nil {
			err = errors.Errorf("cannot pull account balanaces for link without plaid info")
			log.WithError(err).Errorf("failed to pull transactions")
			return err
		}

		bankAccounts, err := repo.GetBankAccountsByLinkId(linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve bank account details to pull transactions")
			return err
		}

		// Gather the plaid account Ids so we can precisely query plaid.
		plaidIdsToBankIds := map[string]uint64{}
		itemBankAccountIds := make([]string, len(bankAccounts))
		for i, bankAccount := range bankAccounts {
			itemBankAccountIds[i] = bankAccount.PlaidAccountId
			plaidIdsToBankIds[bankAccount.PlaidAccountId] = bankAccount.BankAccountId
		}

		log.Debugf("retrieving transactions for %d bank account(s)", len(itemBankAccountIds))

		transactions, err := j.plaidClient.GetAllTransactions(
			span.Context(),
			link.PlaidLink.AccessToken,
			time.Now().Add(-7*24*time.Hour),
			time.Now(),
			itemBankAccountIds,
		)
		if err != nil {
			log.WithError(err).Error("failed to retrieve transactions from plaid")
			return errors.Wrap(err, "failed to retrieve transactions from plaid")
		}

		// TODO Are plaid transaction Ids unique per link, or per bank account?
		//  If they are not then this could cause an issue where a user's checking transaction has the same Id as a
		//  savings account transaction but under the same link. Causing the transaction to get updated improperly.
		plaidTransactionIds := make([]string, len(transactions))
		for i, transaction := range transactions {
			plaidTransactionIds[i] = transaction.ID
		}

		transactionsByPlaidId, err := repo.GetTransactionsByPlaidId(linkId, plaidTransactionIds)
		if err != nil {
			log.WithError(err).Error("failed to retrieve transaction ids for updating plaid transactions")
			return err
		}

		transactionsToUpdate := make([]*models.Transaction, 0)
		transactionsToInsert := make([]models.Transaction, 0)
		now := time.Now().UTC()
		for _, plaidTransaction := range transactions {
			amount := int64(plaidTransaction.Amount * 100)

			date, _ := time.ParseInLocation("2006-01-02", plaidTransaction.Date, timezone)
			var authorizedDate *time.Time
			if plaidTransaction.AuthorizedDate != "" {
				authDate, _ := time.ParseInLocation("2006-01-02", plaidTransaction.AuthorizedDate, timezone)
				authorizedDate = &authDate
			}

			var pendingPlaidTransactionId *string
			if plaidTransaction.PendingTransactionID != "" {
				pendingPlaidTransactionId = &plaidTransaction.PendingTransactionID
			}

			transactionName := plaidTransaction.Name
			if plaidTransaction.MerchantName != "" {
				transactionName = plaidTransaction.MerchantName
			}

			existingTransaction, ok := transactionsByPlaidId[plaidTransaction.ID]
			if !ok {
				transactionsToInsert = append(transactionsToInsert, models.Transaction{
					AccountId:                 accountId,
					BankAccountId:             plaidIdsToBankIds[plaidTransaction.AccountID],
					PlaidTransactionId:        plaidTransaction.ID,
					Amount:                    amount,
					SpendingId:                nil,
					Spending:                  nil,
					Categories:                plaidTransaction.Category,
					OriginalCategories:        plaidTransaction.Category,
					Date:                      date,
					AuthorizedDate:            authorizedDate,
					Name:                      transactionName,
					OriginalName:              plaidTransaction.Name,
					MerchantName:              plaidTransaction.MerchantName,
					OriginalMerchantName:      plaidTransaction.MerchantName,
					IsPending:                 plaidTransaction.Pending,
					CreatedAt:                 now,
					PendingPlaidTransactionId: pendingPlaidTransactionId,
				})
				continue
			}

			var shouldUpdate bool
			if existingTransaction.Amount != amount {
				shouldUpdate = true
			}

			if existingTransaction.IsPending != plaidTransaction.Pending {
				shouldUpdate = true
			}

			if existingTransaction.AuthorizedDate == nil && authorizedDate != nil {
				shouldUpdate = true
			} else if existingTransaction.AuthorizedDate != nil && authorizedDate != nil && !existingTransaction.AuthorizedDate.Equal(*authorizedDate) {
				shouldUpdate = true
			}

			if existingTransaction.PendingPlaidTransactionId != pendingPlaidTransactionId {
				shouldUpdate = true
			}

			existingTransaction.Amount = amount
			existingTransaction.IsPending = plaidTransaction.Pending
			existingTransaction.AuthorizedDate = authorizedDate
			existingTransaction.PendingPlaidTransactionId = pendingPlaidTransactionId

			// Update old records if we see them to use the merchant name by default.
			if existingTransaction.Name == plaidTransaction.Name {
				existingTransaction.Name = transactionName
				shouldUpdate = true
			}

			// Fix timezone of records.
			if existingTransaction.Date != date {
				existingTransaction.Date = date
				shouldUpdate = true
			}

			if shouldUpdate {
				transactionsToUpdate = append(transactionsToUpdate, &existingTransaction)
			}
		}

		if len(transactionsToUpdate) > 0 {
			if err = repo.UpdateTransactions(span.Context(), transactionsToUpdate); err != nil {
				log.WithError(err).Errorf("failed to update transactions for job")
				return err
			}
		}

		if len(transactionsToInsert) > 0 {
			// Reverse the list so the oldest records are inserted first.
			for i, j := 0, len(transactionsToInsert)-1; i < j; i, j = i+1, j-1 {
				transactionsToInsert[i], transactionsToInsert[j] = transactionsToInsert[j], transactionsToInsert[i]
			}
			if err = repo.InsertTransactions(transactionsToInsert); err != nil {
				log.WithError(err).Error("failed to insert new transactions")
				return err
			}
		}

		return nil
	})
}
