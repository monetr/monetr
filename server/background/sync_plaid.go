package background

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	SyncPlaid = "SyncPlaid"
)

var (
	_ ScheduledJobHandler = &SyncPlaidHandler{}
	_ Job                 = &SyncPlaidJob{}
)

type (
	SyncPlaidHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
		publisher     pubsub.Publisher
		enqueuer      JobEnqueuer
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	SyncPlaidArguments struct {
		AccountId uint64 `json:"accountId"`
		LinkId    uint64 `json:"linkId"`
		// Trigger will be "webhook" or "manual"
		Trigger string `json:"trigger"`
	}

	SyncPlaidJob struct {
		args          SyncPlaidArguments
		log           *logrus.Entry
		repo          repository.BaseRepository
		plaidSecrets  secrets.PlaidSecretsProvider
		plaidPlatypus platypus.Platypus
		publisher     pubsub.Publisher
		enqueuer      JobEnqueuer
		clock         clock.Clock
	}
)

func TriggerSyncPlaid(
	ctx context.Context,
	backgroundJobs JobController,
	arguments SyncPlaidArguments,
) error {
	if arguments.Trigger == "" {
		arguments.Trigger = "manual"
	}
	return backgroundJobs.EnqueueJob(ctx, SyncPlaid, arguments)
}

func NewSyncPlaidHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
) *SyncPlaidHandler {
	return &SyncPlaidHandler{
		log:           log,
		db:            db,
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
		publisher:     publisher,
		enqueuer:      enqueuer,
		unmarshaller:  DefaultJobUnmarshaller,
		clock:         clock,
	}
}

func (s SyncPlaidHandler) QueueName() string {
	return SyncPlaid
}

func (s *SyncPlaidHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args SyncPlaidArguments
	if err := errors.Wrap(s.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Sync Plaid job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return s.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		repo := repository.NewRepositoryFromSession(s.clock, 0, args.AccountId, txn)
		job, err := NewSyncPlaidJob(
			s.log.WithContext(span.Context()),
			repo,
			s.clock,
			s.plaidSecrets,
			s.plaidPlatypus,
			s.publisher,
			s.enqueuer,
			args,
		)
		if err != nil {
			return err
		}
		return job.Run(span.Context())
	})
}

func (s SyncPlaidHandler) DefaultSchedule() string {
	// Run every 12 hours. Links that have not received any updates in the last 13 hours will be synced with Plaid. If no
	// updates have been detected then nothing will happen.
	return "0 0 */12 * * *"
}

func (s *SyncPlaidHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := s.log.WithContext(ctx)

	log.Info("retrieving links to sync with Plaid")

	links := make([]models.Link, 0)
	cutoff := s.clock.Now().Add(-48 * time.Hour)
	err := s.db.ModelContext(ctx, &links).
		Join(`INNER JOIN "plaid_links" AS "plaid_link"`).
		JoinOn(`"plaid_link"."plaid_link_id" = "link"."plaid_link_id"`).
		Where(`"plaid_link"."use_plaid_sync" = ?`, true).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Where(`"link"."link_status" = ?`, models.PlaidLinkStatusSetup).
		Where(`"link"."last_attempted_update" < ?`, cutoff).
		Select(&links)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve links that need to by synced with plaid")
	}

	if len(links) == 0 {
		log.Debug("no plaid links need to be synced at this time")
		return nil
	}

	log.WithField("count", len(links)).Info("syncing plaid links")

	for _, item := range links {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
			"linkId":    item.LinkId,
		})
		itemLog.Trace("enqueuing link to be synced with plaid")
		err := enqueuer.EnqueueJob(ctx, s.QueueName(), SyncPlaidArguments{
			AccountId: item.AccountId,
			LinkId:    item.LinkId,
			Trigger:   "cron",
		})
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job to sync with plaid")
			crumbs.Warn(ctx, "Failed to enqueue job to sync with plaid", "job", map[string]interface{}{
				"error": err,
			})
			continue
		}

		itemLog.Trace("successfully enqueued link to be synced with plaid")
	}

	return nil
}

func NewSyncPlaidJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	plaidSecrets secrets.PlaidSecretsProvider,
	plaidPlatypus platypus.Platypus,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
	args SyncPlaidArguments,
) (*SyncPlaidJob, error) {
	return &SyncPlaidJob{
		args:          args,
		log:           log,
		repo:          repo,
		plaidSecrets:  plaidSecrets,
		plaidPlatypus: plaidPlatypus,
		publisher:     publisher,
		enqueuer:      enqueuer,
		clock:         clock,
	}, nil
}

func (s *SyncPlaidJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := s.log.WithContext(span.Context())

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

	if len(bankAccounts) == 0 {
		log.Warn("no bank accounts for plaid link")
		crumbs.Debug(span.Context(), "No bank accounts setup for plaid link", nil)
		return nil
	}

	accessToken, err := s.plaidSecrets.GetAccessTokenForPlaidLinkId(
		span.Context(),
		s.args.AccountId,
		plaidLink.PlaidId,
	)
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

	// now := s.clock.Now().UTC()
	plaidClient, err := s.plaidPlatypus.NewClient(
		span.Context(),
		link,
		accessToken,
		plaidLink.PlaidId,
	)
	if err != nil {
		log.WithError(err).Error("failed to create plaid client for link")
		return err
	}

	plaidBankAccounts, err := plaidClient.GetAccounts(
		span.Context(),
	)
	if err != nil {
		log.WithError(err).Error("failed to retrieve bank accounts from plaid")
		return errors.Wrap(err, "failed to retrieve bank accounts from plaid")
	}

	if len(plaidBankAccounts) == 0 {
		log.Warn("no bank accounts returned by plaid, nothing to sync?")
		crumbs.IndicateBug(span.Context(), "no bank accounts were returned from plaid", nil)
		return nil
	}

	// plaidBankToLocalBank := map[string]models.BankAccount{}
	// bankAccountIds := make([]string, 0, len(bankAccounts))
	//
	// for x := range bankAccounts {
	// 	bankAccount := bankAccounts[x]
	// 	for y := range plaidBankAccounts {
	// 		plaidBankAccount := plaidBankAccounts[y]
	// 		if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
	// 			bankAccountIds = append(bankAccountIds, bankAccount.PlaidBankAccount.PlaidId)
	// 			plaidBankToLocalBank[bankAccount.PlaidBankAccount.PlaidId] = bankAccount
	// 			break
	// 		}
	// 	}
	//
	// 	// If an account is no longer visible in plaid that means that we won't receive updates for that account anymore. If
	// 	// this happens, log something and mark that account as inactive. This way we can inform the user that the account
	// 	// is no longer receiving updates.
	// 	if _, ok := plaidBankToLocalBank[bankAccount.PlaidBankAccount.PlaidId]; !ok {
	// 		log.WithFields(logrus.Fields{
	// 			"bankAccountId": bankAccount.BankAccountId,
	// 		}).Info("found bank account that is no longer present in plaid, it will be updated as inactive")
	// 		crumbs.Warn(span.Context(), "Found bank account that is no longer present in Plaid", "plaid", map[string]interface{}{
	// 			"bankAccountId": bankAccount.BankAccountId,
	// 		})
	// 		bankAccount.Status = models.InactiveBankAccountStatus
	// 		if err = s.repo.UpdateBankAccounts(span.Context(), bankAccount); err != nil {
	// 			log.WithFields(logrus.Fields{
	// 				"bankAccountId": bankAccount.BankAccountId,
	// 			}).
	// 				WithError(err).
	// 				Error("failed to update bank account as inactive")
	// 		}
	// 	}
	// }
	//
	// if len(bankAccountIds) == 0 {
	// 	log.Warn("none of the linked bank accounts are active at plaid")
	// 	crumbs.IndicateBug(span.Context(), "none of the linked bank accounts are active at plaid", nil)
	// 	return nil
	// }
	//
	// crumbs.Debug(span.Context(), "pulling transactions for bank accounts", map[string]interface{}{
	// 	"plaidAccountIds": bankAccountIds,
	// })
	//
	// lastSync, err := s.repo.GetLastPlaidSync(span.Context(), *link.PlaidLinkId)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to retrieve details about previous plaid sync")
	// }
	//
	// var cursor *string
	// if lastSync != nil {
	// 	cursor = &lastSync.NextCursor
	// }
	//
	// transactionSimilaritySyncs := map[uint64]CalculateTransactionClustersArguments{}
	//
	// plaidLink := link.PlaidLink
	// for iter := 0; iter < 10; iter++ {
	// 	syncData, err := plaidClient.Sync(span.Context(), cursor)
	// 	if err != nil {
	// 		return errors.Wrap(err, "failed to sync with plaid")
	// 	}
	//
	// 	// If we received nothing to insert/update/remove then do nothing
	// 	if len(syncData.New)+len(syncData.Updated)+len(syncData.Deleted) == 0 {
	// 		plaidLink.LastAttemptedUpdate = myownsanity.TimeP(s.clock.Now().UTC())
	// 		if err = s.repo.UpdatePlaidLink(span.Context(), plaidLink); err != nil {
	// 			log.WithError(err).Error("failed to update link with last attempt timestamp")
	// 			return err
	// 		}
	//
	// 		log.Info("no new data from plaid, nothing to be done")
	// 		return nil
	// 	}
	//
	// 	// If we did receive something then log that and process it below.
	// 	if err = s.repo.RecordPlaidSync(
	// 		span.Context(),
	// 		*link.PlaidLinkId,
	// 		syncData.NextCursor,
	// 		s.args.Trigger,
	// 		len(syncData.New),
	// 		len(syncData.Updated),
	// 		len(syncData.Deleted),
	// 	); err != nil {
	// 		return errors.Wrap(err, "failed to record plaid sync progress")
	// 	}
	//
	// 	// Update the cursor incase we need to iterate again.
	// 	cursor = &syncData.NextCursor
	//
	// 	plaidTransactions := append(syncData.New, syncData.Updated...)
	//
	// 	log.WithField("count", len(plaidTransactions)).Debugf("retrieved transactions from plaid")
	// 	crumbs.Debug(span.Context(), "Retrieved transactions from plaid.", map[string]interface{}{
	// 		"count": len(plaidTransactions),
	// 	})
	//
	// 	account, err := s.repo.GetAccount(span.Context())
	// 	if err != nil {
	// 		log.WithError(err).Error("failed to retrieve account for job")
	// 		return err
	// 	}
	//
	// 	timezone, err := account.GetTimezone()
	// 	if err != nil {
	// 		log.WithError(err).Warn("failed to get account's time zone, defaulting to UTC")
	// 		timezone = time.UTC
	// 	}
	//
	// 	plaidTransactionIds := make([]string, len(plaidTransactions))
	// 	for i, transaction := range plaidTransactions {
	// 		plaidTransactionIds[i] = transaction.GetTransactionId()
	// 	}
	//
	// 	transactionsByPlaidId, err := s.repo.GetTransactionsByPlaidId(span.Context(), link.LinkId, plaidTransactionIds)
	// 	if err != nil {
	// 		log.WithError(err).Error("failed to retrieve transaction ids for updating plaid transactions")
	// 		return err
	// 	}
	//
	// 	log.Debugf("found %d existing transactions", len(transactionsByPlaidId))
	//
	// 	transactionsToUpdate := make([]*models.Transaction, 0)
	// 	transactionsToInsert := make([]models.Transaction, 0)
	// 	for i := range plaidTransactions {
	// 		plaidTransaction := plaidTransactions[i]
	// 		amount := plaidTransaction.GetAmount()
	//
	// 		date := plaidTransaction.GetDateLocal(timezone)
	//
	// 		transactionName := plaidTransaction.GetName()
	//
	// 		// We only want to make the transaction name be the merchant name if the merchant name is shorter. This is
	// 		// due to something I observed with a dominos transaction, where the merchant was improperly parsed and the
	// 		// transaction ended up being called `Mnuslindstrom` rather than `Domino's`. This should fix that problem.
	// 		if plaidTransaction.GetMerchantName() != "" && len(plaidTransaction.GetMerchantName()) < len(transactionName) {
	// 			transactionName = plaidTransaction.GetMerchantName()
	// 		}
	//
	// 		// TODO this conflicts with the existing check below inside !ok because
	// 		// this might include pending or non pending or both.
	// 		existingTransaction, ok := transactionsByPlaidId[plaidTransaction.GetTransactionId()]
	// 		if !ok {
	// 			accountId := s.repo.AccountId()
	// 			bankAccount := plaidBankToLocalBank[plaidTransaction.GetBankAccountId()]
	//
	// 			// Make sure that we sync for this bank account. This is potentially
	// 			// done multiple times but thats okay because we are tagging these in a
	// 			// map so it is de-dupped by the end.
	// 			transactionSimilaritySyncs[bankAccount.BankAccountId] = CalculateTransactionClustersArguments{
	// 				AccountId:     accountId,
	// 				BankAccountId: bankAccount.BankAccountId,
	// 			}
	//
	// 			pTransaction := models.PlaidTransaction{
	// 				AccountId:          accountId,
	// 				PlaidLinkId:        *link.PlaidLinkId,
	// 				PlaidBankAccountId: *bankAccount.PlaidBankAccountId,
	// 				PlaidId:            plaidTransaction.GetTransactionId(),
	// 				PendingPlaidId:     plaidTransaction.GetPendingTransactionId(),
	// 				Categories:         plaidTransaction.GetCategory(),
	// 				Date:               date,
	// 				Name:               transactionName,
	// 				MerchantName:       plaidTransaction.GetMerchantName(),
	// 				Amount:             amount,
	// 				Currency:           plaidTransaction.GetISOCurrencyCode(),
	// 				IsPending:          plaidTransaction.GetIsPending(),
	// 			}
	// 			if err := s.repo.CreatePlaidTransaction(span.Context(), &pTransaction); err != nil {
	// 				log.WithFields(logrus.Fields{
	// 					"plaidId": plaidTransaction.GetTransactionId(),
	// 				}).
	// 					WithError(err).
	// 					Error("failed to create plaid transaction from sync")
	// 				return errors.Wrap(err, "failed to create plaid transaction")
	// 			}
	//
	// 			mTransaction := &models.Transaction{
	// 				AccountId:     accountId,
	// 				BankAccountId: bankAccount.BankAccountId,
	// 				Amount:        amount,
	// 				SpendingId:    nil,
	// 				Spending:      nil,
	// 				Categories:    plaidTransaction.GetCategory(),
	// 				Date:          date,
	// 				Name:          transactionName,
	// 				MerchantName:  plaidTransaction.GetMerchantName(),
	// 				Currency:      plaidTransaction.GetISOCurrencyCode(),
	// 				IsPending:     plaidTransaction.GetIsPending(),
	// 			}
	//
	// 			// If we are pending then we for sure are creating a new transaction.
	// 			if plaidTransaction.GetIsPending() {
	// 				mTransaction.PendingPlaidTransactionId = &pTransaction.PlaidTransactionId
	// 				transactionsToInsert = append(transactionsToInsert, *mTransaction)
	// 				continue
	// 			}
	//
	// 			// If we have a pending transaction ID then we should try to resolve it
	// 			// to an existing transaction.
	// 			if pendingId := plaidTransaction.GetPendingTransactionId(); pendingId != nil {
	// 				// Overwrite the current transaction object
	// 				existing, err := s.repo.GetTransactionByPendingTransactionPlaidId(
	// 					span.Context(),
	// 					bankAccount.BankAccountId,
	// 					*pendingId,
	// 				)
	// 				if err != nil {
	// 					return errors.Wrap(err, "failed to find pending transaction for sync")
	// 				}
	// 				if existing != nil {
	// 					// Then update the non-pending plaid transaction ID.
	// 					existing.PlaidTransactionId = &pTransaction.PlaidTransactionId
	// 					transactionsToUpdate = append(transactionsToUpdate, existing)
	// 					continue
	// 				}
	// 			}
	//
	// 			// We aren't pending and we don't have an existing pending transaction,
	// 			// just use the model we built and create it.
	// 			mTransaction.PlaidTransactionId = &pTransaction.PlaidTransactionId
	// 			transactionsToInsert = append(transactionsToInsert, *mTransaction)
	// 		}
	//
	// 		var shouldUpdate bool
	// 		if existingTransaction.Amount != amount {
	// 			shouldUpdate = true
	// 		}
	//
	// 		if existingTransaction.IsPending != plaidTransaction.GetIsPending() {
	// 			shouldUpdate = true
	// 		}
	//
	// 		if !myownsanity.StringPEqual(existingTransaction.PendingPlaidTransactionId, plaidTransaction.GetPendingTransactionId()) {
	// 			shouldUpdate = true
	// 		}
	//
	// 		existingTransaction.Amount = amount
	// 		existingTransaction.IsPending = plaidTransaction.GetIsPending()
	// 		existingTransaction.PendingPlaidTransactionId = plaidTransaction.GetPendingTransactionId()
	//
	// 		// Fix timezone of records.
	// 		if !existingTransaction.Date.Equal(date) {
	// 			existingTransaction.Date = date
	// 			shouldUpdate = true
	// 		}
	//
	// 		if shouldUpdate {
	// 			transactionsToUpdate = append(transactionsToUpdate, &existingTransaction)
	// 			transactionSimilaritySyncs[existingTransaction.BankAccountId] = CalculateTransactionClustersArguments{
	// 				AccountId:     existingTransaction.AccountId,
	// 				BankAccountId: existingTransaction.BankAccountId,
	// 			}
	// 		}
	// 	}
	//
	// 	if len(transactionsToUpdate) > 0 {
	// 		log.Infof("updating %d transactions", len(transactionsToUpdate))
	// 		crumbs.Debug(span.Context(), "Updating transactions.", map[string]interface{}{
	// 			"count": len(transactionsToUpdate),
	// 		})
	// 		if err = s.repo.UpdateTransactions(span.Context(), transactionsToUpdate); err != nil {
	// 			log.WithError(err).Errorf("failed to update transactions for job")
	// 			return err
	// 		}
	// 	}
	//
	// 	if len(transactionsToInsert) > 0 {
	// 		// Reverse the list so the oldest records are inserted first.
	// 		for i, j := 0, len(transactionsToInsert)-1; i < j; i, j = i+1, j-1 {
	// 			transactionsToInsert[i], transactionsToInsert[j] = transactionsToInsert[j], transactionsToInsert[i]
	// 		}
	// 		log.Infof("creating %d transactions", len(transactionsToInsert))
	// 		crumbs.Debug(span.Context(), "Creating transactions.", map[string]interface{}{
	// 			"count": len(transactionsToInsert),
	// 		})
	// 		if err = s.repo.InsertTransactions(span.Context(), transactionsToInsert); err != nil {
	// 			log.WithError(err).Error("failed to insert new transactions")
	// 			return err
	// 		}
	// 	}
	//
	// 	if len(transactionsToInsert)+len(transactionsToUpdate) > 0 {
	// 		updatedBankAccounts := make([]models.BankAccount, 0, len(plaidBankAccounts))
	// 		for _, item := range plaidBankAccounts {
	// 			bankAccount, ok := plaidBankToLocalBank[item.GetAccountId()]
	// 			if !ok {
	// 				log.WithField("plaidBankAccountId", item.GetAccountId()).Warn("bank was not found in map")
	// 				continue
	// 			}
	//
	// 			bankLog := log.WithFields(logrus.Fields{
	// 				"bankAccountId": bankAccount.BankAccountId,
	// 				"linkId":        bankAccount.LinkId,
	// 			})
	// 			shouldUpdate := false
	// 			available := item.GetBalances().GetAvailable()
	// 			current := item.GetBalances().GetCurrent()
	//
	// 			if bankAccount.CurrentBalance != current {
	// 				bankLog = bankLog.WithField("currentBalanceChanged", true)
	// 				shouldUpdate = true
	// 			} else {
	// 				bankLog = bankLog.WithField("currentBalanceChanged", false)
	// 			}
	//
	// 			if bankAccount.AvailableBalance != available {
	// 				bankLog = bankLog.WithField("availableBalanceChanged", true)
	// 				shouldUpdate = true
	// 			} else {
	// 				bankLog = bankLog.WithField("availableBalanceChanged", false)
	// 			}
	//
	// 			plaidName := bankAccount.PlaidName
	// 			if bankAccount.PlaidName != item.GetName() {
	// 				plaidName = item.GetName()
	// 				shouldUpdate = true
	// 				bankLog = bankLog.WithField("plaidNameChanged", true)
	// 			} else {
	// 				bankLog = bankLog.WithField("plaidNameChanged", false)
	// 			}
	//
	// 			plaidOfficialName := bankAccount.PlaidOfficialName
	// 			if bankAccount.PlaidOfficialName != item.GetOfficialName() {
	// 				plaidOfficialName = item.GetOfficialName()
	// 				shouldUpdate = true
	// 				bankLog = bankLog.WithField("plaidOfficialNameChanged", true)
	// 			} else {
	// 				bankLog = bankLog.WithField("plaidOfficialNameChanged", false)
	// 			}
	//
	// 			bankLog = bankLog.WithField("willUpdate", shouldUpdate)
	//
	// 			if shouldUpdate {
	// 				bankLog.Info("updating bank account balances")
	// 			} else {
	// 				bankLog.Trace("balances do not need to be updated")
	// 			}
	//
	// 			if shouldUpdate {
	// 				updatedBankAccounts = append(updatedBankAccounts, models.BankAccount{
	// 					BankAccountId:     bankAccount.BankAccountId,
	// 					AccountId:         s.args.AccountId,
	// 					AvailableBalance:  available,
	// 					CurrentBalance:    current,
	// 					PlaidName:         plaidName,
	// 					PlaidOfficialName: plaidOfficialName,
	// 					LastUpdated:       now.UTC(),
	// 				})
	// 			}
	// 		}
	//
	// 		if err = s.repo.UpdateBankAccounts(span.Context(), updatedBankAccounts...); err != nil {
	// 			log.WithError(err).Error("failed to update bank account balances")
	// 			crumbs.ReportError(span.Context(), err, "Failed to update bank account balances", "job", nil)
	// 		}
	// 	}
	//
	// 	if len(syncData.Deleted) > 0 { // Handle removed transactions
	// 		log.Infof("removing %d transaction(s)", len(syncData.Deleted))
	//
	// 		transactions, err := s.repo.GetTransactionsByPlaidTransactionId(span.Context(), s.args.LinkId, syncData.Deleted)
	// 		if err != nil {
	// 			log.WithError(err).Error("failed to retrieve transactions by plaid transaction Id for removal")
	// 			return err
	// 		}
	//
	// 		if len(transactions) == 0 {
	// 			log.Warnf("no transactions retrieved, nothing to be done. transactions might already have been deleted")
	// 			return nil
	// 		}
	//
	// 		if len(transactions) != len(syncData.Deleted) {
	// 			log.Warnf("number of transactions retrieved does not match expected number of transactions, expected: %d found: %d", len(syncData.Deleted), len(transactions))
	// 			crumbs.IndicateBug(span.Context(), "The number of transactions retrieved does not match the expected number of transactions", map[string]interface{}{
	// 				"expected":            len(syncData.Deleted),
	// 				"found":               len(transactions),
	// 				"plaidTransactionIds": syncData.Deleted,
	// 			})
	// 		}
	//
	// 		for _, existingTransaction := range transactions {
	// 			transactionSimilaritySyncs[existingTransaction.BankAccountId] = CalculateTransactionClustersArguments{
	// 				AccountId:     existingTransaction.AccountId,
	// 				BankAccountId: existingTransaction.BankAccountId,
	// 			}
	//
	// 			if existingTransaction.SpendingId == nil {
	// 				continue
	// 			}
	//
	// 			// If the transaction is spent from something then we need to remove the spent from before deleting it to
	// 			// maintain our balances correctly.
	// 			updatedTransaction := existingTransaction
	// 			updatedTransaction.SpendingId = nil
	//
	// 			// This is a simple sanity check, working with objects in slices and for loops can be goofy, or my
	// 			// understanding of the way objects works with how they are referenced in memory is poor. This is to make
	// 			// sure im not doing it wrong though. I'm worried that making a "copy" of the object and then modifying the
	// 			// copy will modify the original as well.
	// 			if existingTransaction.SpendingId == nil {
	// 				sentry.CaptureMessage("original transaction modified")
	// 				panic("original transaction modified")
	// 			}
	//
	// 			_, err = s.repo.ProcessTransactionSpentFrom(
	// 				span.Context(),
	// 				existingTransaction.BankAccountId,
	// 				&updatedTransaction,
	// 				&existingTransaction,
	// 			)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	//
	// 		for _, transaction := range transactions {
	// 			if err = s.repo.DeleteTransaction(span.Context(), transaction.BankAccountId, transaction.TransactionId); err != nil {
	// 				log.WithField("transactionId", transaction.TransactionId).WithError(err).
	// 					Error("failed to delete transaction")
	// 				return err
	// 			}
	// 		}
	//
	// 		log.Debugf("successfully removed %d transaction(s)", len(transactions))
	// 	}
	//
	// 	if !syncData.HasMore {
	// 		break
	// 	}
	//
	// 	log.WithField("iter", iter).Info("there is more data to sync from plaid, continuing")
	// }
	//
	// linkWasSetup := false
	//
	// // If the link status is not setup or pending expiration. Then change the status to setup
	// switch link.LinkStatus {
	// case models.PlaidLinkStatusSetup, models.PlaidLinkStatusPendingExpiration:
	// default:
	// 	crumbs.Debug(span.Context(), "Updating link status.", map[string]interface{}{
	// 		"old": link.LinkStatus,
	// 		"new": models.PlaidLinkStatusSetup,
	// 	})
	// 	link.LinkStatus = models.PlaidLinkStatusSetup
	// 	linkWasSetup = true
	// }
	// link.LastSuccessfulUpdate = myownsanity.TimeP(s.clock.Now().UTC())
	// link.LastAttemptedUpdate = myownsanity.TimeP(s.clock.Now().UTC())
	// if err = s.repo.UpdateLink(span.Context(), link); err != nil {
	// 	log.WithError(err).Error("failed to update link after transaction sync")
	// 	return err
	// }
	//
	// // Then enqueue all of the bank accounts we touched to have their similar
	// // transactions recalculated.
	// for key := range transactionSimilaritySyncs {
	// 	s.enqueuer.EnqueueJob(span.Context(), CalculateTransactionClusters, transactionSimilaritySyncs[key])
	// }
	//
	// if linkWasSetup { // Send the notification that the link has been set up.
	// 	channelName := fmt.Sprintf("initial:plaid:link:%d:%d", s.args.AccountId, s.args.LinkId)
	// 	if notifyErr := s.publisher.Notify(
	// 		span.Context(),
	// 		channelName,
	// 		"success",
	// 	); notifyErr != nil {
	// 		log.WithError(notifyErr).Error("failed to publish link status to pubsub")
	// 	}
	// }

	// TODO Trigger similar transaction calculation here.
	return nil
}
