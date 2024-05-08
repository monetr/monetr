package background

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
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
	_ JobImplementation   = &SyncPlaidJob{}
)

type (
	SyncPlaidHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		kms           secrets.KeyManagement
		plaidPlatypus platypus.Platypus
		publisher     pubsub.Publisher
		enqueuer      JobEnqueuer
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	SyncPlaidArguments struct {
		AccountId ID[Account] `json:"accountId"`
		LinkId    ID[Link]    `json:"linkId"`
		// Trigger will be "webhook" or "manual" or "command"
		Trigger string `json:"trigger"`
	}

	SyncPlaidJob struct {
		args          SyncPlaidArguments
		log           *logrus.Entry
		repo          repository.BaseRepository
		secrets       repository.SecretsRepository
		plaidPlatypus platypus.Platypus
		publisher     pubsub.Publisher
		enqueuer      JobEnqueuer
		clock         clock.Clock

		timezone     *time.Location
		bankAccounts map[string]BankAccount
		transactions map[string]Transaction
		similarity   map[ID[BankAccount]]CalculateTransactionClustersArguments
		actions      map[ID[Transaction]]SyncAction
	}

	SyncChange struct {
		Field string `json:"field"`
		Old   any    `json:"old"`
		New   any    `json:"new"`
	}

	SyncAction string
)

const (
	CreateSyncAction SyncAction = "create"
	UpdateSyncAction SyncAction = "update"
	DeleteSyncAction SyncAction = "delete"
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
	kms secrets.KeyManagement,
	plaidPlatypus platypus.Platypus,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
) *SyncPlaidHandler {
	return &SyncPlaidHandler{
		log:           log,
		db:            db,
		kms:           kms,
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

		log := s.log.WithContext(span.Context())

		repo := repository.NewRepositoryFromSession(s.clock, "user_plaid", args.AccountId, txn)
		secretsRepo := repository.NewSecretsRepository(
			log,
			s.clock,
			txn,
			s.kms,
			args.AccountId,
		)
		job, err := NewSyncPlaidJob(
			log,
			repo,
			s.clock,
			secretsRepo,
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

	links := make([]Link, 0)
	cutoff := s.clock.Now().Add(-48 * time.Hour)
	err := s.db.ModelContext(ctx, &links).
		Join(`INNER JOIN "plaid_links" AS "plaid_link"`).
		JoinOn(`"plaid_link"."plaid_link_id" = "link"."plaid_link_id"`).
		Where(`"plaid_link"."status" = ?`, PlaidLinkStatusSetup).
		Where(`"plaid_link"."last_attempted_update" < ?`, cutoff).
		Where(`"link"."link_type" = ?`, PlaidLinkType).
		Where(`"link"."deleted_at" IS NULL`).
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
	secrets repository.SecretsRepository,
	plaidPlatypus platypus.Platypus,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
	args SyncPlaidArguments,
) (*SyncPlaidJob, error) {
	return &SyncPlaidJob{
		args:          args,
		log:           log,
		repo:          repo,
		secrets:       secrets,
		plaidPlatypus: plaidPlatypus,
		publisher:     publisher,
		enqueuer:      enqueuer,
		clock:         clock,

		timezone:     nil, // Is set below
		transactions: make(map[string]Transaction),
		bankAccounts: make(map[string]BankAccount),
		similarity:   make(map[ID[BankAccount]]CalculateTransactionClustersArguments),
		actions:      make(map[ID[Transaction]]SyncAction),
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

	account, err := s.repo.GetAccount(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to retrieve account for job")
		return err
	}

	s.timezone, err = account.GetTimezone()
	if err != nil {
		log.WithError(err).Warn("failed to get account's time zone, defaulting to UTC")
		s.timezone = time.UTC
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

	for x := range bankAccounts {
		bankAccount := bankAccounts[x]
		for y := range plaidBankAccounts {
			plaidBankAccount := plaidBankAccounts[y]
			if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
				s.bankAccounts[bankAccount.PlaidBankAccount.PlaidId] = bankAccount
				break
			}
		}

		// If an account is no longer visible in plaid that means that we won't receive updates for that account anymore. If
		// this happens, log something and mark that account as inactive. This way we can inform the user that the account
		// is no longer receiving updates.
		if _, ok := s.bankAccounts[bankAccount.PlaidBankAccount.PlaidId]; !ok {
			log.WithFields(logrus.Fields{
				"bankAccountId": bankAccount.BankAccountId,
			}).Info("found bank account that is no longer present in plaid, it will be updated as inactive")
			crumbs.Warn(span.Context(), "Found bank account that is no longer present in Plaid", "plaid", map[string]interface{}{
				"bankAccountId": bankAccount.BankAccountId,
			})
			if err := s.syncPlaidBankAccount(
				span.Context(),
				link,
				&bankAccount,
				plaidLink,
				bankAccount.PlaidBankAccount,
				nil, // Not visible via sync anymore
			); err != nil {
				log.WithFields(logrus.Fields{
					"bankAccountId": bankAccount.BankAccountId,
				}).
					WithError(err).
					Error("failed to update bank account as inactive")
			}
		}
	}

	if len(s.bankAccounts) == 0 {
		log.Warn("none of the linked bank accounts are active at plaid")
		crumbs.IndicateBug(span.Context(), "none of the linked bank accounts are active at plaid", nil)
		return nil
	}

	lastSync, err := s.repo.GetLastPlaidSync(span.Context(), *link.PlaidLinkId)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve details about previous plaid sync")
	}

	var cursor *string
	if lastSync != nil {
		cursor = &lastSync.NextCursor
	}

	for iter := 0; iter < 10; iter++ {
		syncData, err := plaidClient.Sync(span.Context(), cursor)
		if err != nil {
			return errors.Wrap(err, "failed to sync with plaid")
		}

		// If we received nothing to insert/update/remove then do nothing
		if len(syncData.New)+len(syncData.Updated)+len(syncData.Deleted) == 0 {
			plaidLink.LastAttemptedUpdate = myownsanity.TimeP(s.clock.Now().UTC())
			if err = s.repo.UpdatePlaidLink(span.Context(), plaidLink); err != nil {
				log.WithError(err).Error("failed to update link with last attempt timestamp")
				return err
			}

			log.Info("no new data from plaid, nothing to be done")
			return nil
		}

		// If we did receive something then log that and process it below.
		if err = s.repo.RecordPlaidSync(
			span.Context(),
			*link.PlaidLinkId,
			syncData.NextCursor,
			s.args.Trigger,
			len(syncData.New),
			len(syncData.Updated),
			len(syncData.Deleted),
		); err != nil {
			return errors.Wrap(err, "failed to record plaid sync progress")
		}

		// Update the cursor incase we need to iterate again.
		cursor = &syncData.NextCursor

		plaidTransactions := append(syncData.New, syncData.Updated...)

		log.WithField("count", len(plaidTransactions)).Debugf("retrieved transactions from plaid")
		crumbs.Debug(span.Context(), "Retrieved transactions from plaid.", map[string]interface{}{
			"count": len(plaidTransactions),
		})

		if err := s.hydrateTransactions(span.Context(), link, syncData); err != nil {
			return errors.Wrap(err, "failed to hydrate existing transaction data")
		}

		log.Debugf("found %d existing transactions", len(s.transactions))

		transactionsToUpdate := make([]*Transaction, 0)
		transactionsToInsert := make([]Transaction, 0)
		for i := range plaidTransactions {
			plaidTransaction := plaidTransactions[i]
			bankAccount := s.bankAccounts[plaidTransaction.GetBankAccountId()]

			created, updated, err := s.syncPlaidTransaction(
				span.Context(),
				link,
				&bankAccount,
				plaidLink,
				bankAccount.PlaidBankAccount,
				plaidTransaction,
			)
			if err != nil {
				return errors.Wrap(err, "failed to sync transaction")
			}

			if created != nil {
				transactionsToInsert = append(transactionsToInsert, *created)
				s.tagBankAccountForSimilarityRecalc(bankAccount.BankAccountId)
			} else if updated != nil {
				transactionsToUpdate = append(transactionsToUpdate, updated)
				s.tagBankAccountForSimilarityRecalc(bankAccount.BankAccountId)
			}

			continue
		}

		if len(transactionsToUpdate) > 0 {
			log.Infof("updating %d transactions", len(transactionsToUpdate))
			crumbs.Debug(span.Context(), "Updating transactions.", map[string]interface{}{
				"count": len(transactionsToUpdate),
			})
			if err = s.repo.UpdateTransactions(span.Context(), transactionsToUpdate); err != nil {
				log.WithError(err).Errorf("failed to update transactions for job")
				return err
			}
			for i := range transactionsToUpdate {
				s.actions[transactionsToUpdate[i].TransactionId] = UpdateSyncAction
			}
		}

		if len(transactionsToInsert) > 0 {
			// Sort by oldest to newest
			sort.Slice(transactionsToInsert, func(i, j int) bool {
				return transactionsToInsert[i].Date.Before(transactionsToInsert[j].Date)
			})

			// // Reverse the list so the oldest records are inserted first.
			// for i, j := 0, len(transactionsToInsert)-1; i < j; i, j = i+1, j-1 {
			// 	transactionsToInsert[i], transactionsToInsert[j] = transactionsToInsert[j], transactionsToInsert[i]
			// }
			log.Infof("creating %d transactions", len(transactionsToInsert))
			crumbs.Debug(span.Context(), "Creating transactions.", map[string]interface{}{
				"count": len(transactionsToInsert),
			})
			if err = s.repo.InsertTransactions(span.Context(), transactionsToInsert); err != nil {
				log.WithError(err).Error("failed to insert new transactions")
				return err
			}
			for i := range transactionsToInsert {
				s.actions[transactionsToInsert[i].TransactionId] = CreateSyncAction
			}
		}

		for _, item := range plaidBankAccounts {
			bankAccount, ok := s.bankAccounts[item.GetAccountId()]
			if !ok {
				log.WithField("plaidBankAccountId", item.GetAccountId()).Warn("bank was not found in map")
				continue
			}

			if err := s.syncPlaidBankAccount(
				span.Context(),
				link,
				&bankAccount,
				plaidLink,
				bankAccount.PlaidBankAccount,
				item,
			); err != nil {
				log.WithError(err).Error("failed to update bank account")
				crumbs.ReportError(span.Context(), err, "Failed to update bank account", "job", nil)
			}
		}

		// Handle deleted transactions
		for i := range syncData.Deleted {
			if err := s.syncRemovedTransaction(
				ctx,
				link,
				plaidLink,
				syncData.Deleted[i],
			); err != nil {
				return errors.Wrap(err, "failed to sync deleted transaction")
			}
		}

		if !syncData.HasMore {
			break
		}

		log.WithField("iter", iter).Info("there is more data to sync from plaid, continuing")
	}

	// Then enqueue all of the bank accounts we touched to have their similar
	// transactions recalculated.
	for key := range s.similarity {
		s.enqueuer.EnqueueJob(span.Context(), CalculateTransactionClusters, s.similarity[key])
	}

	return s.maintainLinkStatus(ctx, plaidLink)
}

func (s *SyncPlaidJob) tagBankAccountForSimilarityRecalc(bankAccountId ID[BankAccount]) {
	s.similarity[bankAccountId] = CalculateTransactionClustersArguments{
		AccountId:     s.args.AccountId,
		BankAccountId: bankAccountId,
	}
}

func (s *SyncPlaidJob) maintainLinkStatus(ctx context.Context, plaidLink *PlaidLink) error {
	linkWasSetup := false
	// If the link status is not setup or pending expiration. Then change the status to setup
	switch plaidLink.Status {
	case PlaidLinkStatusSetup, PlaidLinkStatusPendingExpiration:
	default:
		crumbs.Debug(ctx, "Updating plaid link status.", map[string]interface{}{
			"old": plaidLink.Status,
			"new": PlaidLinkStatusSetup,
		})
		plaidLink.Status = PlaidLinkStatusSetup
		linkWasSetup = true
	}
	plaidLink.LastSuccessfulUpdate = myownsanity.TimeP(s.clock.Now().UTC())
	plaidLink.LastAttemptedUpdate = myownsanity.TimeP(s.clock.Now().UTC())
	if err := s.repo.UpdatePlaidLink(ctx, plaidLink); err != nil {
		s.log.WithError(err).Error("failed to update link after transaction sync")
		return err
	}

	if linkWasSetup { // Send the notification that the link has been set up.
		channelName := fmt.Sprintf("initial:plaid:link:%s:%s", s.args.AccountId, s.args.LinkId)
		if notifyErr := s.publisher.Notify(
			ctx,
			channelName,
			"success",
		); notifyErr != nil {
			s.log.WithError(notifyErr).Error("failed to publish link status to pubsub")
		}
	}

	return nil
}

// hydrateTransactions takes all of the transaction's retrieved from Plaid
// (including deleted ones please) and retrieves them and stores them on the job
// object. This way when we are processing the transactions we can calculate
// differences between the transactions retrieved and the ones we have stored.
func (s *SyncPlaidJob) hydrateTransactions(
	ctx context.Context,
	link *Link,
	sync *platypus.SyncResult,
) error {
	plaidTransactionIds := make([]string, 0, len(sync.Deleted)+len(sync.Updated)+len(sync.New))
	for _, transaction := range sync.New {
		plaidTransactionIds = append(plaidTransactionIds, transaction.GetTransactionId())
	}
	for _, transaction := range sync.Updated {
		plaidTransactionIds = append(plaidTransactionIds, transaction.GetTransactionId())
	}
	plaidTransactionIds = append(plaidTransactionIds, sync.Deleted...)

	s.log.
		WithContext(ctx).
		Tracef("checking database for %d plaid transaction(s)", len(plaidTransactionIds))

	var err error
	s.transactions, err = s.repo.GetTransactionsByPlaidId(
		ctx,
		link.LinkId,
		plaidTransactionIds,
	)
	if err != nil {
		s.log.
			WithContext(ctx).
			WithError(err).
			Error("failed to retrieve transaction ids for updating plaid transactions")
		return err
	}

	return nil
}

func (s *SyncPlaidJob) lookupTransaction(
	plaidId string,
	pendingPlaidId *string,
) (Transaction, bool) {
	txn, ok := s.transactions[plaidId]
	if ok {
		return txn, ok
	}
	if pendingPlaidId != nil {
		txn, ok = s.transactions[*pendingPlaidId]
		return txn, ok
	}

	return Transaction{}, false
}

func (s *SyncPlaidJob) syncPlaidTransaction(
	ctx context.Context,
	link *Link,
	bankAccount *BankAccount,
	plaidLink *PlaidLink,
	plaidBankAccount *PlaidBankAccount,
	input platypus.Transaction,
) (created, updated *Transaction, err error) {
	existingTransaction, exists := s.lookupTransaction(
		input.GetTransactionId(),
		input.GetPendingTransactionId(),
	)

	amount := input.GetAmount()
	date := input.GetDateLocal(s.timezone)
	transactionName := input.GetName()

	// We only want to make the transaction name be the merchant name if the
	// merchant name is shorter. This is due to something I observed with a
	// dominos transaction, where the merchant was improperly parsed and the
	// transaction ended up being called `Mnuslindstrom` rather than `Domino's`.
	// This should fix that problem.
	if input.GetMerchantName() != "" && len(input.GetMerchantName()) < len(transactionName) {
		transactionName = input.GetMerchantName()
	}

	if !exists {
		plaidTransaction := PlaidTransaction{
			AccountId:          link.AccountId,
			PlaidBankAccountId: plaidBankAccount.PlaidBankAccountId,
			PlaidId:            input.GetTransactionId(),
			PendingPlaidId:     input.GetPendingTransactionId(),
			Categories:         input.GetCategory(),
			Date:               date,
			Name:               transactionName,
			MerchantName:       input.GetMerchantName(),
			Amount:             amount,
			Currency:           input.GetISOCurrencyCode(),
			IsPending:          input.GetIsPending(),
		}
		if err := s.repo.CreatePlaidTransaction(ctx, &plaidTransaction); err != nil {
			return nil, nil, errors.Wrap(err, "failed to store new plaid transaction")
		}

		existingTransaction = Transaction{
			AccountId:            link.AccountId,
			BankAccountId:        bankAccount.BankAccountId,
			Amount:               amount,
			SpendingId:           nil,
			SpendingAmount:       nil,
			Categories:           input.GetCategory(),
			Date:                 date,
			Name:                 transactionName,
			OriginalName:         input.GetName(),
			MerchantName:         input.GetMerchantName(),
			OriginalMerchantName: input.GetMerchantName(),
			Currency:             input.GetISOCurrencyCode(),
			IsPending:            input.GetIsPending(),
		}

		if input.GetIsPending() {
			existingTransaction.PendingPlaidTransactionId = &plaidTransaction.PlaidTransactionId
		} else {
			existingTransaction.PlaidTransactionId = &plaidTransaction.PlaidTransactionId
		}

		return &existingTransaction, nil, nil
	}

	var existingPlaidTransaction *PlaidTransaction
	if input.GetIsPending() {
		existingPlaidTransaction = existingTransaction.PendingPlaidTransaction
	} else {
		existingPlaidTransaction = existingTransaction.PlaidTransaction
	}

	if existingPlaidTransaction == nil && input.GetIsPending() {
		crumbs.IndicateBug(ctx, "Existing transaction did not correctly have the associated pending plaid transaction stored", map[string]interface{}{
			"plaidId":            input.GetTransactionId(),
			"linkId":             link.LinkId,
			"plaidLinkId":        link.PlaidLinkId,
			"bankAccountId":      bankAccount.BankAccountId,
			"plaidBankAccountId": bankAccount.PlaidBankAccountId,
			"institutionId":      plaidLink.InstitutionId,
			"itemId":             plaidLink.PlaidId,
		})
		panic("existing plaid transaction is missing, there is a bug")
	}

	changes := make([]SyncChange, 0)

	// If the existing plaid transaction is nil and we are not pending that means
	// we have transitioned from a pending status to a cleared status for this
	// transaction. We need to create the new plaid transaction for this input.
	if existingPlaidTransaction == nil {
		existingPlaidTransaction = &PlaidTransaction{
			AccountId:          link.AccountId,
			PlaidBankAccountId: plaidBankAccount.PlaidBankAccountId,
			PlaidId:            input.GetTransactionId(),
			PendingPlaidId:     input.GetPendingTransactionId(),
			Categories:         input.GetCategory(),
			Date:               date,
			Name:               transactionName,
			MerchantName:       input.GetMerchantName(),
			Amount:             amount,
			Currency:           input.GetISOCurrencyCode(),
			IsPending:          input.GetIsPending(),
		}
		if err := s.repo.CreatePlaidTransaction(ctx, existingPlaidTransaction); err != nil {
			return nil, nil, errors.Wrap(err, "failed to store new plaid transaction")
		}

		existingTransaction.PlaidTransactionId = &existingPlaidTransaction.PlaidTransactionId
		changes = append(changes, SyncChange{
			Field: "plaidTransactionId",
			Old:   nil,
			New:   existingPlaidTransaction.PlaidTransactionId,
		})
	}

	if existingPlaidTransaction.Amount != existingTransaction.Amount {
		changes = append(changes, SyncChange{
			Field: "amount",
			Old:   existingTransaction.Amount,
			New:   existingPlaidTransaction.Amount,
		})
		existingTransaction.Amount = existingPlaidTransaction.Amount
	}

	if existingPlaidTransaction.Date != existingTransaction.Date {
		changes = append(changes, SyncChange{
			Field: "date",
			Old:   existingTransaction.Date,
			New:   existingPlaidTransaction.Date,
		})
		existingTransaction.Date = existingPlaidTransaction.Date
	}

	if existingPlaidTransaction.Name != existingTransaction.Name {
		changes = append(changes, SyncChange{
			Field: "name",
			Old:   existingTransaction.Name,
			New:   existingPlaidTransaction.Name,
		})
		existingTransaction.Name = existingPlaidTransaction.Name
	}

	if existingPlaidTransaction.MerchantName != existingTransaction.MerchantName {
		changes = append(changes, SyncChange{
			Field: "merchantName",
			Old:   existingTransaction.MerchantName,
			New:   existingPlaidTransaction.MerchantName,
		})
		existingTransaction.MerchantName = existingPlaidTransaction.MerchantName
	}

	if existingPlaidTransaction.IsPending != existingTransaction.IsPending {
		changes = append(changes, SyncChange{
			Field: "isPending",
			Old:   existingTransaction.IsPending,
			New:   existingPlaidTransaction.IsPending,
		})
		existingTransaction.IsPending = existingPlaidTransaction.IsPending
	}

	// This happens when a transactions that is pending has it's pending
	// transaction removed (the pending is not visible anymore). But the
	// non-pending transaction has not appeared yet. Then when the non-pending
	// transaction becomes visible (sometime later) this happens and we have to
	// undelete the transaction.
	if existingPlaidTransaction.DeletedAt == nil && existingTransaction.DeletedAt != nil {
		changes = append(changes, SyncChange{
			Field: "deletedAt",
			Old:   existingTransaction.DeletedAt,
			New:   nil,
		})
		existingTransaction.DeletedAt = nil
	}

	// If any of the fields did change, log the changes and return the updated
	// transaction object.
	if len(changes) > 0 {
		s.log.WithContext(ctx).WithFields(logrus.Fields{
			"plaidId": input.GetTransactionId(),
			"kind":    "transaction",
			"changes": changes,
		}).Debug("detected transaction updates from plaid")
		return nil, &existingTransaction, nil
	}

	// There were no changes but no errors.
	return nil, nil, nil
}

func (s *SyncPlaidJob) syncRemovedTransaction(
	ctx context.Context,
	link *Link,
	plaidLink *PlaidLink,
	id string,
) error {
	log := s.log.WithFields(logrus.Fields{
		"itemId":  plaidLink.PlaidId,
		"linkId":  link.LinkId,
		"kind":    "transaction",
		"plaidId": id,
	})
	existingTransaction, exists := s.lookupTransaction(id, &id)
	if !exists {
		log.Warn("plaid wants to remove a transaction that does not exist")
		return nil
	}
	log = log.WithFields(logrus.Fields{
		"bankAccountId":             existingTransaction.BankAccountId,
		"transactionId":             existingTransaction.TransactionId,
		"plaidTransactionId":        existingTransaction.PlaidTransactionId,
		"pendingPlaidTransactionId": existingTransaction.PendingPlaidTransactionId,
	})

	action := s.actions[existingTransaction.TransactionId]
	switch action {
	// TODO At the moment Created would not actually be detected.
	case CreateSyncAction, UpdateSyncAction:
		// If a transaction was updated or created as part of this sync then that
		// means the transaction we are deleting was likely a pending transaction
		// and the cleared transaction has become available and was properly
		// associated with the pending transaction in Plaid. As such we should not
		// remove the transaction since it should have the correct status now.
		// TODO Keep an eye on this, the logic is new and might be wrong.
		log.WithField("action", action).Debug("transaction to be removed has also been created or updated in this sync, it will not be removed")
	default:
		s.tagBankAccountForSimilarityRecalc(existingTransaction.BankAccountId)

		log.Debug("removing transaction")

		if existingTransaction.SpendingId != nil {
			log.WithField("spendingId", existingTransaction.SpendingId).
				Debug("transaction has spending, it will be removed")
			updatedTransaction := existingTransaction
			updatedTransaction.SpendingId = nil
			_, err := s.repo.ProcessTransactionSpentFrom(
				ctx,
				existingTransaction.BankAccountId,
				&updatedTransaction,
				&existingTransaction,
			)
			if err != nil {
				return err
			}
		}

		// Safe to remove this transaction
		if err := s.repo.DeleteTransaction(
			ctx,
			existingTransaction.BankAccountId,
			existingTransaction.TransactionId,
		); err != nil {
			return errors.Wrap(err, "failed to remove pending transaction")
		}
	}

	return nil
}

func (s *SyncPlaidJob) syncPlaidBankAccount(
	ctx context.Context,
	link *Link,
	bankAccount *BankAccount,
	plaidLink *PlaidLink,
	plaidBankAccount *PlaidBankAccount,
	input platypus.BankAccount,
) error {
	// If input is nil that means we are no longer seeing this specific account
	// and we should mark it as inactive.
	if input == nil {
		// TODO, should we add a similar status to the plaid bank account?
		bankAccount.Status = InactiveBankAccountStatus
		bankAccount.LastUpdated = s.clock.Now().UTC()
		return s.repo.UpdateBankAccounts(ctx, *bankAccount)
	}

	changes := make([]SyncChange, 0)
	if input.GetName() != plaidBankAccount.Name {
		changes = append(changes, SyncChange{
			Field: "name",
			Old:   plaidBankAccount.Name,
			New:   input.GetName(),
		})
		plaidBankAccount.Name = input.GetName()
		bankAccount.OriginalName = input.GetName()
	}

	if input.GetBalances().GetAvailable() != bankAccount.AvailableBalance {
		changes = append(changes, SyncChange{
			Field: "availableBalance",
			Old:   bankAccount.AvailableBalance,
			New:   input.GetBalances().GetAvailable(),
		})
		plaidBankAccount.AvailableBalance = input.GetBalances().GetAvailable()
		bankAccount.AvailableBalance = input.GetBalances().GetAvailable()
	}

	if input.GetBalances().GetCurrent() != bankAccount.CurrentBalance {
		changes = append(changes, SyncChange{
			Field: "currentBalance",
			Old:   bankAccount.CurrentBalance,
			New:   input.GetBalances().GetCurrent(),
		})
		plaidBankAccount.CurrentBalance = input.GetBalances().GetCurrent()
		bankAccount.CurrentBalance = input.GetBalances().GetCurrent()
	}

	if len(changes) > 0 {
		s.log.WithContext(ctx).WithFields(logrus.Fields{
			"plaidId": input.GetAccountId(),
			"kind":    "bankAccount",
			"changes": changes,
		}).Debug("detected bank account updates from plaid")

		if err := s.repo.UpdateBankAccounts(ctx, *bankAccount); err != nil {
			return errors.Wrap(err, "failed to persists bank account changes from plaid sync")
		}

		if err := s.repo.UpdatePlaidBankAccount(ctx, plaidBankAccount); err != nil {
			return errors.Wrap(err, "failed to persists plaid bank account changes from plaid sync")
		}
	}

	return nil
}
