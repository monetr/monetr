package jobs

import (
	"fmt"
	"github.com/gocraft/work"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"time"
)

const (
	PullInitialTransactions = "PullInitialTransactions"
)

func (j *jobManagerBase) pullInitialTransactions(job *work.Job) error {
	log := j.getLogForJob(job)

	accountId := uint64(job.ArgInt64("accountId"))

	linkId := uint64(job.ArgInt64("linkId"))
	if linkId == 0 {
		log.Error("cannot pull initial transactions without a link Id")
		return errors.Errorf("cannot pull initial transactions without a link Id")
	}

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(linkId)
		if err != nil {
			log.WithError(err).Error("cannot pull initial transactions for link provided")
			return nil
		}

		if link.PlaidLink == nil {
			log.Error("provided link does not have any plaid credentials")
			return nil
		}

		if len(link.BankAccounts) == 0 {
			log.Error("no bank accounts for plaid link")
			return nil
		}

		bankAccountIdsByPlaid := map[string]uint64{}
		bankAccountIds := make([]string, len(link.BankAccounts))
		for i, bankAccount := range link.BankAccounts {
			bankAccountIds[i] = bankAccount.PlaidAccountId
			bankAccountIdsByPlaid[bankAccount.PlaidAccountId] = bankAccount.BankAccountId
		}

		plaidTransactions := make([]plaid.Transaction, 0)
		var offset int
		for {
			log.WithField("offset", offset).Debug("retrieving transactions from plaid")
			response, err := j.plaidClient.GetTransactionsWithOptions(link.PlaidLink.AccessToken, plaid.GetTransactionsOptions{
				StartDate:  time.Now().Add(-2 * 365 * 24 * time.Hour).Format("2006-01-02"),
				EndDate:    time.Now().Format("2006-01-02"),
				AccountIDs: bankAccountIds,
				Count:      500,
				Offset:     offset,
			})
			if err != nil {
				log.WithError(err).Error("failed to retrieve some transactions from plaid")
				break
			}

			plaidTransactions = append(plaidTransactions, response.Transactions...)

			if len(response.Transactions) < 500 {
				break
			}
		}

		if len(plaidTransactions) == 0 {
			log.Warn("no transactions were retrieved from plaid")
			return nil
		}

		log.Debugf("retreived %d transaction(s) from plaid, processing now", len(plaidTransactions))

		now := time.Now().UTC()
		transactions := make([]models.Transaction, len(plaidTransactions))
		for i, plaidTransaction := range plaidTransactions {
			date, _ := time.Parse("2006-01-02", plaidTransaction.Date)
			var authorizedDate *time.Time
			if plaidTransaction.AuthorizedDate != "" {
				authDate, _ := time.Parse("2006-01-02", plaidTransaction.AuthorizedDate)
				authorizedDate = &authDate
			}
			transactions[i] = models.Transaction{
				AccountId:            repo.AccountId(),
				BankAccountId:        bankAccountIdsByPlaid[plaidTransaction.AccountID],
				PlaidTransactionId:   plaidTransaction.ID,
				Amount:               int64(plaidTransaction.Amount * 100),
				ExpenseId:            nil,
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
			}
		}

		// Reverse the list so the oldest records are inserted first.
		for i, j := 0, len(transactions)-1; i < j; i, j = i+1, j-1 {
			transactions[i], transactions[j] = transactions[j], transactions[i]
		}

		if err := repo.InsertTransactions(transactions); err != nil {
			log.WithError(err).Error("failed to store initial transactions")
			return err
		}

		_, err = j.db.Exec(fmt.Sprintf(`NOTIFY job_%d_%s, ?`, accountId, job.ID), "DONE")
		if err != nil {
			return err
		}
		return nil
	})
}
