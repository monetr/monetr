package jobs

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/internal/platypus"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/sirupsen/logrus"
)

func (j *jobManagerBase) upsertTransactions(
	ctx context.Context,
	log *logrus.Entry,
	repo repository.BaseRepository,
	link *models.Link,
	plaidIdsToBankIds map[string]uint64,
	plaidTransactions []platypus.Transaction,
) error {
	span := sentry.StartSpan(ctx, "Job - Upsert Transactions")
	defer span.Finish()

	account, err := repo.GetAccount(span.Context())
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

	transactionsByPlaidId, err := repo.GetTransactionsByPlaidId(span.Context(), link.LinkId, plaidTransactionIds)
	if err != nil {
		log.WithError(err).Error("failed to retrieve transaction ids for updating plaid transactions")
		return err
	}

	transactionsToUpdate := make([]*models.Transaction, 0)
	transactionsToInsert := make([]models.Transaction, 0)
	now := time.Now().UTC()
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
				AccountId:                 repo.AccountId(),
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
		log.Infof("updating %d transactions", len(transactionsToUpdate))
		if err = repo.UpdateTransactions(span.Context(), transactionsToUpdate); err != nil {
			log.WithError(err).Errorf("failed to update transactions for job")
			return err
		}
	}

	if len(transactionsToInsert) > 0 {
		log.Infof("creating %d transactions", len(transactionsToInsert))
		// Reverse the list so the oldest records are inserted first.
		for i, j := 0, len(transactionsToInsert)-1; i < j; i, j = i+1, j-1 {
			transactionsToInsert[i], transactionsToInsert[j] = transactionsToInsert[j], transactionsToInsert[i]
		}
		if err = repo.InsertTransactions(span.Context(), transactionsToInsert); err != nil {
			log.WithError(err).Error("failed to insert new transactions")
			return err
		}
	}

	return nil
}
