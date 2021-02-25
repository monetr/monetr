package jobs

import (
	"github.com/gocraft/work"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	EnqueuePullLatestTransactions = "EnqueuePullLatestTransactions"
	PullLatestTransactions        = "PullLatestTransactions"
)

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

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(linkId)
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

		log.Tracef("retrieving transactions for %d bank account(s)", len(itemBankAccountIds))

		transactions := make([]plaid.Transaction, 0)
		perPage := 200
		for {
			result, err := j.plaidClient.GetTransactionsWithOptions(
				link.PlaidLink.AccessToken,
				plaid.GetTransactionsOptions{
					// TODO How do we want to determine our pull latest transactions window.
					StartDate:  time.Now().Add(-7 * 24 * time.Hour).Format("2006-01-02"),
					EndDate:    time.Now().Format("2006-01-02"),
					AccountIDs: itemBankAccountIds,
					Count:      perPage,
					Offset:     len(transactions),
				},
			)
			if err != nil {
				log.WithError(err).Error("failed to retrieve transactions from plaid")
				return errors.Wrap(err, "failed to retrieve transactions from plaid")
			}

			transactions = append(transactions, result.Transactions...)

			// If we get a page with fewer transactions than we requested, that means we have reached the end.
			if len(result.Transactions) < perPage {
				break
			}
		}

		// TODO Are plaid transaction Ids unique per link, or per bank account?
		//  If they are not then this could cause an issue where a user's checking transaction has the same Id as a
		//  savings account transaction but under the same link. Causing the transaction to get updated improperly.
		plaidTransactionIds := make([]string, len(transactions))
		for i, transaction := range transactions {
			plaidTransactionIds[i] = transaction.ID
		}

		transactionIds, err := repo.GetTransactionsByPlaidId(linkId, plaidTransactionIds)
		if err != nil {
			log.WithError(err).Error("failed to retrieve transaction ids for updating plaid transactions")
			return err
		}

		transactionsToUpdate := make([]models.Transaction, 0)
		transactionsToInsert := make([]models.Transaction, 0)
		now := time.Now().UTC()
		for _, plaidTransaction := range transactions {
			amount := int64(plaidTransaction.Amount * 100)

			existingTransaction, ok := transactionIds[plaidTransaction.ID]
			if !ok {
				date, _ := time.Parse("2006-01-02", plaidTransaction.Date)
				var authorizedDate *time.Time
				if plaidTransaction.AuthorizedDate != "" {
					authDate, _ := time.Parse("2006-01-02", plaidTransaction.AuthorizedDate)
					authorizedDate = &authDate
				}
				transactionsToInsert = append(transactionsToInsert, models.Transaction{
					AccountId:            accountId,
					BankAccountId:        plaidIdsToBankIds[plaidTransaction.AccountID],
					PlaidTransactionId:   plaidTransaction.ID,
					Amount:               amount,
					ExpenseId:            nil,
					Expense:              nil,
					Categories:           plaidTransaction.Category,
					OriginalCategories:   plaidTransaction.Category,
					Date:                 date,
					AuthorizedDate:       authorizedDate,
					Name:                 plaidTransaction.Name,
					OriginalName:         plaidTransaction.Name,
					MerchantName:         plaidTransaction.MerchantName,
					OriginalMerchantName: plaidTransaction.MerchantName,
					IsPending:            plaidTransaction.Pending,
					CreatedAt:            now,
				})
				continue
			}

			if amount == existingTransaction.Amount {
				continue
			}

			transactionsToUpdate = append(transactionsToUpdate, models.Transaction{
				TransactionId:      existingTransaction.TransactionId,
				AccountId:          accountId,
				BankAccountId:      existingTransaction.BankAccountId,
				PlaidTransactionId: plaidTransaction.ID,
				Amount:             amount,
				IsPending:          plaidTransaction.Pending,
			})
		}

		if len(transactionsToUpdate) > 0 {
			// Update the transactions
		}

		if len(transactionsToInsert) > 0 {
			if err = repo.InsertTransactions(transactionsToInsert); err != nil {
				log.WithError(err).Error("failed to insert new transactions")
				return err
			}
		}

		return nil
	})
}
