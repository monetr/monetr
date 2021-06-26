package jobs

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/monetr/rest-api/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	PullHistoricalTransactions = "PullHistoricalTransactions"
)

func (j *jobManagerBase) TriggerPullHistoricalTransactions(accountId, linkId uint64) (jobId string, err error) {
	log := j.log.WithFields(logrus.Fields{
		"accountId": accountId,
		"linkId":    linkId,
	})

	log.Infof("queueing pull historical transactions for account")
	job, err := j.queue.EnqueueUnique(PullHistoricalTransactions, map[string]interface{}{
		"accountId": accountId,
		"linkId":    linkId,
	})
	if err != nil {
		log.WithError(err).Error("failed to enqueue pulling historical transactions")
		return "", errors.Wrap(err, "failed to enqueue pulling historical transactions")
	}

	return job.ID, nil
}

func (j *jobManagerBase) pullHistoricalTransactions(job *work.Job) error {
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Pull Historical Transactions"))
	defer span.Finish()

	log := j.getLogForJob(job)
	log.Infof("pulling historical transactions")

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	linkId := uint64(job.ArgInt64("linkId"))
	span.SetTag("linkId", strconv.FormatUint(linkId, 10))
	span.SetTag("accountId", strconv.FormatUint(accountId, 10))

	twoYearsAgo := time.Now().Add(-2 * 365 * 24 * time.Hour).UTC()

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
			log.WithError(err).Error("failed to retrieve link details to pull historical transactions")
			return err
		}

		if link.PlaidLink == nil {
			err = errors.Errorf("cannot pull account balanaces for link without plaid info")
			log.WithError(err).Errorf("failed to pull transactions")
			return err
		}

		accessToken, err := j.plaidSecrets.GetAccessTokenForPlaidLinkId(span.Context(), accountId, link.PlaidLink.ItemId)
		if err != nil {
			log.WithError(err).Errorf("failed to retrieve access token for link")
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
			accessToken,
			twoYearsAgo,
			time.Now(),
			itemBankAccountIds,
		)
		if err != nil {
			log.WithError(err).Error("failed to retrieve transactions from plaid")
			return errors.Wrap(err, "failed to retrieve transactions from plaid")
		}

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

			date, _ := util.ParseInLocal("2006-01-02", plaidTransaction.Date, timezone)
			var authorizedDate *time.Time
			if plaidTransaction.AuthorizedDate != "" {
				authDate, _ := util.ParseInLocal("2006-01-02", plaidTransaction.AuthorizedDate, timezone)
				authorizedDate = &authDate
			}

			var pendingPlaidTransactionId *string
			if plaidTransaction.PendingTransactionID != "" {
				pendingPlaidTransactionId = &plaidTransaction.PendingTransactionID
			}

			transactionName := plaidTransaction.Name

			// We only want to make the transaction name be the merchant name if the merchant name is shorter. This is
			// due to something I observed with a dominos transaction, where the merchant was improperly parsed and the
			// transaction ended up being called `Mnuslindstrom` rather than `Domino's`. This should fix that problem.
			if plaidTransaction.MerchantName != "" && len(plaidTransaction.MerchantName) < len(transactionName) {
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

			if !myownsanity.TimesPEqual(existingTransaction.AuthorizedDate, authorizedDate) {
				shouldUpdate = true
			}

			if existingTransaction.PendingPlaidTransactionId != pendingPlaidTransactionId {
				shouldUpdate = true
			}

			existingTransaction.Amount = amount
			existingTransaction.IsPending = plaidTransaction.Pending
			existingTransaction.AuthorizedDate = authorizedDate
			existingTransaction.PendingPlaidTransactionId = pendingPlaidTransactionId

			// Update old transactions calculated name as we can.
			if existingTransaction.Name != transactionName {
				existingTransaction.Name = transactionName
				shouldUpdate = true
			}

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
			if err = repo.InsertTransactions(span.Context(), transactionsToInsert); err != nil {
				log.WithError(err).Error("failed to insert new transactions")
				return err
			}
		}

		link.LastSuccessfulUpdate = myownsanity.TimeP(time.Now().UTC())
		return repo.UpdateLink(link)
	})
}
