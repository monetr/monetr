package plaid_jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/similar/similar_jobs"
	"github.com/pkg/errors"
)

type SyncAction string

const (
	CreateSyncAction SyncAction = "create"
	UpdateSyncAction SyncAction = "update"
	DeleteSyncAction SyncAction = "delete"
)

type SyncChange struct {
	Field string `json:"field"`
	Old   any    `json:"old"`
	New   any    `json:"new"`
}

type SyncPlaidArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
	LinkId    models.ID[models.Link]    `json:"linkId"`
	// Trigger will be "webhook" or "manual" or "command"
	Trigger string `json:"trigger"`
}

type syncPlaidJob struct {
	args    SyncPlaidArguments
	log     *slog.Logger
	repo    repository.BaseRepository
	secrets repository.SecretsRepository

	timezone     *time.Location
	bankAccounts map[string]models.BankAccount
	transactions map[string]models.Transaction
	similarity   map[models.ID[models.BankAccount]]similar_jobs.CalculateTransactionClustersArguments
	actions      map[models.ID[models.Transaction]]SyncAction
}

func (s *syncPlaidJob) lookupTransaction(
	plaidId string,
	pendingPlaidId *string,
) (models.Transaction, bool) {
	txn, ok := s.transactions[plaidId]
	if ok {
		return txn, ok
	}
	if pendingPlaidId != nil {
		txn, ok = s.transactions[*pendingPlaidId]
		return txn, ok
	}

	return models.Transaction{}, false
}

func (s *syncPlaidJob) tagBankAccountForSimilarityRecalc(
	bankAccountId models.ID[models.BankAccount],
) {
	s.similarity[bankAccountId] = similar_jobs.CalculateTransactionClustersArguments{
		AccountId:     s.args.AccountId,
		BankAccountId: bankAccountId,
	}
}

// hydrateTransactions takes all of the transaction's retrieved from Plaid
// (including deleted ones please) and retrieves them and stores them on the job
// object. This way when we are processing the transactions we can calculate
// differences between the transactions retrieved and the ones we have stored.
func (s *syncPlaidJob) hydrateTransactions(
	ctx context.Context,
	link *models.Link,
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

	s.log.Log(
		ctx,
		logging.LevelTrace,
		"checking database for plaid transactions",
		"count", len(plaidTransactionIds),
	)

	var err error
	s.transactions, err = s.repo.GetTransactionsByPlaidId(
		ctx,
		link.LinkId,
		plaidTransactionIds,
	)
	if err != nil {
		s.log.ErrorContext(
			ctx,
			"failed to retrieve transaction ids for updating plaid transactions",
			"err", err,
		)
		return err
	}

	return nil
}

func (s *syncPlaidJob) syncPlaidTransaction(
	ctx context.Context,
	link *models.Link,
	bankAccount *models.BankAccount,
	plaidLink *models.PlaidLink,
	plaidBankAccount *models.PlaidBankAccount,
	input platypus.Transaction,
) (created, updated *models.Transaction, plaidCreated *models.PlaidTransaction, err error) {
	existingTransaction, exists := s.lookupTransaction(
		input.GetTransactionId(),
		input.GetPendingTransactionId(),
	)

	amount := input.GetAmount()
	date := input.GetDateLocal(s.timezone).UTC()
	transactionName := input.GetName()

	// We only want to make the transaction name be the merchant name if the
	// merchant name is shorter. This is due to something I observed with a
	// dominos transaction, where the merchant was improperly parsed and the
	// transaction ended up being called `Mnuslindstrom` rather than `Domino's`.
	// This should fix that problem.
	if input.GetMerchantName() != "" && len(input.GetMerchantName()) < len(transactionName) {
		transactionName = input.GetMerchantName()
	}

	originalName := input.GetOriginalDescription()
	if originalName == "" {
		originalName = transactionName
	}

	// If there is not a monetr transaction record for this Plaid transaction,
	// then we simply need to create both records. There is nothing else we need
	// to do here.
	if !exists {
		plaidTransaction := models.PlaidTransaction{
			PlaidTransactionId: models.NewID[models.PlaidTransaction](),
			AccountId:          link.AccountId,
			PlaidBankAccountId: plaidBankAccount.PlaidBankAccountId,
			PlaidId:            input.GetTransactionId(),
			PendingPlaidId:     input.GetPendingTransactionId(),
			Categories:         input.GetCategory(),
			Category:           input.GetCategoryDetail(),
			Date:               date,
			Name:               transactionName,
			MerchantName:       input.GetMerchantName(),
			Amount:             amount,
			Currency:           input.GetISOCurrencyCode(),
			IsPending:          input.GetIsPending(),
		}

		existingTransaction = models.Transaction{
			TransactionId:        models.NewID[models.Transaction](),
			AccountId:            link.AccountId,
			BankAccountId:        bankAccount.BankAccountId,
			Amount:               amount,
			SpendingId:           nil,
			SpendingAmount:       nil,
			Categories:           input.GetCategory(),
			Category:             input.GetCategoryDetail(),
			Date:                 date,
			Name:                 transactionName,
			OriginalName:         originalName,
			MerchantName:         input.GetMerchantName(),
			OriginalMerchantName: input.GetMerchantName(),
			IsPending:            input.GetIsPending(),
			Source:               models.TransactionSourcePlaid,
		}

		if input.GetIsPending() {
			existingTransaction.PendingPlaidTransactionId = &plaidTransaction.PlaidTransactionId
		} else {
			existingTransaction.PlaidTransactionId = &plaidTransaction.PlaidTransactionId
		}

		return &existingTransaction, nil, &plaidTransaction, nil
	}

	// However, if monetr does have a transaction for this plaid transaction; then
	// we have to establish whether or not it was a pending transaction and what
	// things have changed since we last saw the transaction.
	var existingPlaidTransaction *models.PlaidTransaction
	if input.GetIsPending() {
		existingPlaidTransaction = existingTransaction.PendingPlaidTransaction
	} else {
		existingPlaidTransaction = existingTransaction.PlaidTransaction
	}

	if existingPlaidTransaction == nil && input.GetIsPending() {
		crumbs.IndicateBug(
			ctx,
			"Existing transaction did not correctly have the associated pending plaid transaction stored",
			map[string]any{
				"plaidId":            input.GetTransactionId(),
				"linkId":             link.LinkId,
				"plaidLinkId":        link.PlaidLinkId,
				"bankAccountId":      bankAccount.BankAccountId,
				"plaidBankAccountId": bankAccount.PlaidBankAccountId,
				"institutionId":      plaidLink.InstitutionId,
				"itemId":             plaidLink.PlaidId,
			},
		)
		panic("existing plaid transaction is missing, there is a bug")
	}

	changes := make([]SyncChange, 0)

	// If the existing plaid transaction is nil and we are not pending that means
	// we have transitioned from a pending status to a cleared status for this
	// transaction. We need to create the new plaid transaction for this input.
	create := existingPlaidTransaction == nil
	if existingPlaidTransaction == nil {
		existingPlaidTransaction = &models.PlaidTransaction{
			PlaidTransactionId: models.NewID[models.PlaidTransaction](),
			AccountId:          link.AccountId,
			PlaidBankAccountId: plaidBankAccount.PlaidBankAccountId,
			PlaidId:            input.GetTransactionId(),
			PendingPlaidId:     input.GetPendingTransactionId(),
			Categories:         input.GetCategory(),
			Category:           input.GetCategoryDetail(),
			Date:               date,
			Name:               transactionName,
			MerchantName:       input.GetMerchantName(),
			Amount:             amount,
			Currency:           input.GetISOCurrencyCode(),
			IsPending:          input.GetIsPending(),
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

	if !myownsanity.StringPEqual(existingPlaidTransaction.Category, existingTransaction.Category) {
		changes = append(changes, SyncChange{
			Field: "category",
			Old:   existingTransaction.Category,
			New:   existingPlaidTransaction.Category,
		})
		existingTransaction.Category = existingPlaidTransaction.Category
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
		// Overwrite the original name when the transaction clears.
		// This seems unintuitive but see
		// https://github.com/monetr/monetr/issues/1714 for more information.
		if !existingPlaidTransaction.IsPending {
			existingTransaction.OriginalName = existingPlaidTransaction.Name
		}
	}

	if existingPlaidTransaction.MerchantName != existingTransaction.MerchantName {
		changes = append(changes, SyncChange{
			Field: "merchantName",
			Old:   existingTransaction.MerchantName,
			New:   existingPlaidTransaction.MerchantName,
		})
		existingTransaction.MerchantName = existingPlaidTransaction.MerchantName
		// Same as above, overwrite if we aren't pending.
		if !existingPlaidTransaction.IsPending {
			existingTransaction.OriginalMerchantName = existingPlaidTransaction.MerchantName
		}
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
		s.log.DebugContext(ctx, "detected transaction updates from plaid",
			"plaidId", input.GetTransactionId(),
			"kind", "transaction",
			"changes", changes,
		)
		if create {
			return nil, &existingTransaction, existingPlaidTransaction, nil
		} else {
			return nil, &existingTransaction, nil, nil
		}
	}

	// There were no changes but no errors.
	if create {
		return nil, nil, existingPlaidTransaction, nil
	} else {
		return nil, nil, nil, nil
	}
}

func (s *syncPlaidJob) syncRemovedTransaction(
	ctx context.Context,
	link *models.Link,
	plaidLink *models.PlaidLink,
	id string,
) error {
	log := s.log.With(
		"itemId", plaidLink.PlaidId,
		"linkId", link.LinkId,
		"kind", "transaction",
		"plaidId", id,
	)
	existingTransaction, exists := s.lookupTransaction(id, &id)
	if !exists {
		log.WarnContext(ctx, "plaid wants to remove a transaction that does not exist")
		return nil
	}
	log = log.With(
		"bankAccountId", existingTransaction.BankAccountId,
		"transactionId", existingTransaction.TransactionId,
		"plaidTransactionId", existingTransaction.PlaidTransactionId,
		"pendingPlaidTransactionId", existingTransaction.PendingPlaidTransactionId,
	)

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
		log.DebugContext(ctx, "transaction to be removed has also been created or updated in this sync, it will not be removed", "action", action)
	default:
		s.tagBankAccountForSimilarityRecalc(existingTransaction.BankAccountId)

		log.DebugContext(ctx, "removing transaction")

		if existingTransaction.SpendingId != nil {
			log.DebugContext(ctx, "transaction has spending, it will be removed", "spendingId", existingTransaction.SpendingId)
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
		if err := s.repo.SoftDeleteTransaction(
			ctx,
			existingTransaction.BankAccountId,
			existingTransaction.TransactionId,
		); err != nil {
			return errors.Wrap(err, "failed to remove pending transaction")
		}
	}

	return nil
}

func (s *syncPlaidJob) syncPlaidBankAccount(
	ctx queue.Context,
	link *models.Link,
	bankAccount *models.BankAccount,
	plaidLink *models.PlaidLink,
	plaidBankAccount *models.PlaidBankAccount,
	input platypus.BankAccount,
) error {
	changes := make([]SyncChange, 0)

	// If input is nil that means we are no longer seeing this specific account
	// and we should mark it as inactive.
	if input == nil && bankAccount.Status != models.BankAccountStatusInactive {
		changes = append(changes, SyncChange{
			Field: "status",
			Old:   models.BankAccountStatusActive,
			New:   models.BankAccountStatusInactive,
		})
		bankAccount.Status = models.BankAccountStatusInactive
	}

	// If we observe the account again, then change it back to active.
	if input != nil && bankAccount.Status == models.BankAccountStatusInactive {
		changes = append(changes, SyncChange{
			Field: "status",
			Old:   models.BankAccountStatusInactive,
			New:   models.BankAccountStatusActive,
		})
		bankAccount.Status = models.BankAccountStatusActive
	}

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

	if input.GetBalances().GetLimit() != bankAccount.LimitBalance {
		changes = append(changes, SyncChange{
			Field: "limitBalance",
			Old:   bankAccount.LimitBalance,
			New:   input.GetBalances().GetLimit(),
		})
		plaidBankAccount.LimitBalance = input.GetBalances().GetLimit()
		bankAccount.LimitBalance = input.GetBalances().GetLimit()
	}

	if len(changes) > 0 {
		bankAccount.LastUpdated = ctx.Clock().Now().UTC()
		s.log.DebugContext(ctx, "detected bank account updates from plaid",
			"plaidId", input.GetAccountId(),
			"kind", "bankAccount",
			"changes", changes,
		)

		if err := s.repo.UpdateBankAccount(ctx, bankAccount); err != nil {
			return errors.Wrap(err, "failed to persists bank account changes from plaid sync")
		}

		if err := s.repo.UpdatePlaidBankAccount(ctx, plaidBankAccount); err != nil {
			return errors.Wrap(err, "failed to persists plaid bank account changes from plaid sync")
		}
	}

	return nil
}

func (s *syncPlaidJob) maintainLinkStatus(
	ctx queue.Context,
	plaidLink *models.PlaidLink,
) error {
	linkWasSetup := false
	// If the link status is not setup or pending expiration. Then change the status to setup
	switch plaidLink.Status {
	case models.PlaidLinkStatusSetup, models.PlaidLinkStatusPendingExpiration:
	default:
		crumbs.Debug(ctx, "Updating plaid link status.", map[string]any{
			"old": plaidLink.Status,
			"new": models.PlaidLinkStatusSetup,
		})
		plaidLink.Status = models.PlaidLinkStatusSetup
		linkWasSetup = true
	}
	now := ctx.Clock().Now().UTC()
	plaidLink.LastSuccessfulUpdate = &now
	plaidLink.LastAttemptedUpdate = &now
	if err := s.repo.UpdatePlaidLink(ctx, plaidLink); err != nil {
		s.log.ErrorContext(ctx, "failed to update link after transaction sync", "err", err)
		return err
	}

	if linkWasSetup { // Send the notification that the link has been set up.
		channelName := fmt.Sprintf("initial:plaid:link:%s:%s", s.args.AccountId, s.args.LinkId)
		if notifyErr := ctx.Publisher().Notify(
			ctx,
			s.args.AccountId,
			channelName,
			"success",
		); notifyErr != nil {
			s.log.ErrorContext(ctx, "failed to publish link status to pubsub", "err", notifyErr)
		}
	}

	return nil
}

func SyncPlaidCron(ctx queue.Context) error {
	log := ctx.Log()

	log.InfoContext(ctx, "retrieving links to sync with Plaid")

	links := make([]models.Link, 0)
	cutoff := ctx.Clock().Now().Add(-48 * time.Hour)
	err := ctx.DB().ModelContext(ctx, &links).
		Join(`INNER JOIN "plaid_links" AS "plaid_link"`).
		JoinOn(`"plaid_link"."plaid_link_id" = "link"."plaid_link_id"`).
		Where(`"plaid_link"."status" = ?`, models.PlaidLinkStatusSetup).
		Where(`"plaid_link"."last_attempted_update" < ?`, cutoff).
		Where(`"plaid_link"."deleted_at" IS NULL`).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Where(`"link"."deleted_at" IS NULL`).
		Select(&links)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve links that need to by synced with plaid")
	}

	if len(links) == 0 {
		log.DebugContext(ctx, "no plaid links need to be synced at this time")
		return nil
	}

	log.InfoContext(ctx, "syncing plaid links", "count", len(links))

	for _, item := range links {
		itemLog := log.With(
			"accountId", item.AccountId,
			"linkId", item.LinkId,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing link to be synced with plaid")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			SyncPlaid,
			SyncPlaidArguments{
				AccountId: item.AccountId,
				LinkId:    item.LinkId,
				Trigger:   "cron",
			},
		); err != nil {
			itemLog.WarnContext(ctx, "failed to enqueue job to sync with plaid", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to sync with plaid", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued link to be synced with plaid")
	}

	return nil
}

func SyncPlaid(ctx queue.Context, args SyncPlaidArguments) error {
	crumbs.IncludeUserInScope(ctx, args.AccountId)
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		span := sentry.SpanFromContext(ctx)
		s := &syncPlaidJob{
			args: args,
			log: ctx.Log().With(
				"accountId", args.AccountId,
				"linkId", args.LinkId,
			),
			repo: repository.NewRepositoryFromSession(
				ctx.Clock(),
				"user_plaid",
				args.AccountId,
				ctx.DB(),
				ctx.Log(),
			),
			secrets: repository.NewSecretsRepository(
				ctx.Log().With(
					"accountId", args.AccountId,
					"linkId", args.LinkId,
				),
				ctx.Clock(),
				ctx.DB(),
				ctx.KMS(),
				args.AccountId,
			),
			bankAccounts: map[string]models.BankAccount{},
			transactions: map[string]models.Transaction{},
			similarity:   map[models.ID[models.BankAccount]]similar_jobs.CalculateTransactionClustersArguments{},
			actions:      map[models.ID[models.Transaction]]SyncAction{},
		}

		link, err := s.repo.GetLink(ctx, s.args.LinkId)
		if err = errors.Wrap(err, "failed to retrieve link to sync with plaid"); err != nil {
			s.log.ErrorContext(ctx, "cannot sync without link", "err", err)
			return err
		}

		if link.PlaidLink == nil {
			s.log.WarnContext(ctx, "provided link does not have any plaid credentials")
			crumbs.IndicateBug(
				ctx,
				"BUG: Link was queued to sync with plaid, but has no plaid details",
				map[string]any{
					"link": link,
				},
			)
			span.Status = sentry.SpanStatusFailedPrecondition
			return nil
		}

		s.log = s.log.With(
			"plaidLinkId", link.PlaidLink.PlaidLinkId,
			slog.Group("plaid",
				"institutionId", link.PlaidLink.InstitutionId,
				"institutionName", link.PlaidLink.InstitutionName,
				"itemId", link.PlaidLink.PlaidId,
			),
		)
		account, err := s.repo.GetAccount(ctx)
		if err != nil {
			s.log.ErrorContext(ctx, "failed to retrieve account for job", "err", err)
			return err
		}

		s.timezone, err = account.GetTimezone()
		if err != nil {
			s.log.WarnContext(ctx, "failed to get account's time zone, defaulting to UTC", "err", err)
			s.timezone = time.UTC
		}

		plaidLink := link.PlaidLink

		bankAccounts, err := s.repo.GetBankAccountsWithPlaidByLinkId(
			ctx,
			link.LinkId,
		)
		if err = errors.Wrap(err, "failed to read bank accounts for plaid sync"); err != nil {
			s.log.ErrorContext(ctx, "cannot sync without bank accounts", "err", err)
			return err
		}

		crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.PlaidId)
		crumbs.AddTag(ctx, "plaid.institution_id", link.PlaidLink.InstitutionId)
		crumbs.AddTag(ctx, "plaid.institution_name", link.PlaidLink.InstitutionName)

		if len(bankAccounts) == 0 {
			s.log.WarnContext(ctx, "no bank accounts for plaid link")
			crumbs.Debug(ctx, "No bank accounts setup for plaid link", nil)
			return nil
		}

		secret, err := s.secrets.Read(ctx, plaidLink.SecretId)
		if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
			s.log.ErrorContext(ctx, "could not retrieve API credentials for Plaid for link, this job will be retried", "err", err)
			return err
		}

		plaidClient, err := ctx.Platypus().NewClient(
			ctx,
			link,
			secret.Value,
			plaidLink.PlaidId,
		)
		if err != nil {
			s.log.ErrorContext(ctx, "failed to create plaid client for link", "err", err)
			return err
		}

		// Declare this ahead of the sync below.
		var plaidBankAccounts []platypus.BankAccount

		lastSync, err := s.repo.GetLastPlaidSync(ctx, *link.PlaidLinkId)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve details about previous plaid sync")
		}

		var cursor *string
		if lastSync != nil {
			cursor = &lastSync.NextCursor
		}

		for iter := 0; iter < 10; iter++ {
			syncData, err := plaidClient.Sync(ctx, cursor)
			if err != nil {
				return errors.Wrap(err, "failed to sync with plaid")
			}

			plaidBankAccounts = syncData.Accounts
			for x := range bankAccounts {
				bankAccount := bankAccounts[x]
				for y := range plaidBankAccounts {
					plaidBankAccount := plaidBankAccounts[y]
					if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
						s.bankAccounts[bankAccount.PlaidBankAccount.PlaidId] = bankAccount
						break
					}
				}
			}

			// If we received nothing to insert/update/remove then do nothing
			if len(syncData.New)+len(syncData.Updated)+len(syncData.Deleted) == 0 {
				plaidLink.LastAttemptedUpdate = myownsanity.Pointer(ctx.Clock().Now().UTC())
				if err = s.repo.UpdatePlaidLink(ctx, plaidLink); err != nil {
					s.log.ErrorContext(ctx, "failed to update link with last attempt timestamp", "err", err)
					return err
				}

				s.log.InfoContext(ctx, "no new data from plaid, nothing to be done")
				return nil
			}

			// If we did receive something then log that and process it below.
			if err = s.repo.RecordPlaidSync(
				ctx,
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

			s.log.DebugContext(ctx, fmt.Sprintf("retrieved transactions from plaid"), "count", len(plaidTransactions))
			crumbs.Debug(ctx, "Retrieved transactions from plaid.", map[string]any{
				"count": len(plaidTransactions),
			})

			if err := s.hydrateTransactions(ctx, link, syncData); err != nil {
				return errors.Wrap(err, "failed to hydrate existing transaction data")
			}

			s.log.DebugContext(ctx, fmt.Sprintf("found %d existing transactions", len(s.transactions)))

			transactionsToUpdate := make([]*models.Transaction, 0)
			transactionsToInsert := make([]models.Transaction, 0)
			plaidTransactionsToInsert := make([]*models.PlaidTransaction, 0)
			for i := range plaidTransactions {
				plaidTransaction := plaidTransactions[i]
				bankAccount, ok := s.bankAccounts[plaidTransaction.GetBankAccountId()]
				if !ok {
					s.log.ErrorContext(ctx, "bank account for plaid transaction was not in the bank accounts map! there is a bug!",
						"plaidTransactionId", plaidTransaction.GetTransactionId(),
						"plaidBankAccountId", plaidTransaction.GetBankAccountId(),
						"bug", true,
					)
					crumbs.IndicateBug(
						ctx,
						"bank account for plaid transaction was not in the bank accounts map! there is a bug!",
						map[string]any{
							"plaidTransactionId": plaidTransaction.GetTransactionId(),
							"plaidBankAccountId": plaidTransaction.GetBankAccountId(),
							"bug":                true,
						},
					)
					continue
				}

				created, updated, plaidCreated, err := s.syncPlaidTransaction(
					ctx,
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

				if plaidCreated != nil {
					plaidTransactionsToInsert = append(plaidTransactionsToInsert, plaidCreated)
				}

				continue
			}

			if len(plaidTransactionsToInsert) > 0 {
				s.log.InfoContext(ctx, fmt.Sprintf("creating %d plaid transactions", len(plaidTransactionsToInsert)))
				if err := s.repo.CreatePlaidTransactions(ctx, plaidTransactionsToInsert...); err != nil {
					s.log.ErrorContext(ctx, "failed to create plaid transactions for job", "err", err)
					return err
				}
			}

			if len(transactionsToUpdate) > 0 {
				s.log.InfoContext(ctx, fmt.Sprintf("updating %d transactions", len(transactionsToUpdate)))
				crumbs.Debug(ctx, "Updating transactions.", map[string]any{
					"count": len(transactionsToUpdate),
				})
				if err = s.repo.UpdateTransactions(ctx, transactionsToUpdate); err != nil {
					s.log.ErrorContext(ctx, "failed to update transactions for job", "err", err)
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

				s.log.InfoContext(ctx, fmt.Sprintf("creating %d transactions", len(transactionsToInsert)))
				crumbs.Debug(ctx, "Creating transactions.", map[string]any{
					"count": len(transactionsToInsert),
				})
				if err = s.repo.InsertTransactions(ctx, transactionsToInsert); err != nil {
					s.log.ErrorContext(ctx, "failed to insert new transactions", "err", err)
					return err
				}
				for i := range transactionsToInsert {
					s.actions[transactionsToInsert[i].TransactionId] = CreateSyncAction
				}
			}

			for _, item := range plaidBankAccounts {
				bankAccount, ok := s.bankAccounts[item.GetAccountId()]
				if !ok {
					s.log.WarnContext(ctx, "bank was not found in map", "plaidBankAccountId", item.GetAccountId())
					continue
				}

				if err := s.syncPlaidBankAccount(
					ctx,
					link,
					&bankAccount,
					plaidLink,
					bankAccount.PlaidBankAccount,
					item,
				); err != nil {
					s.log.ErrorContext(ctx, "failed to update bank account", "err", err)
					crumbs.ReportError(ctx, err, "Failed to update bank account", "job", nil)
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

			s.log.InfoContext(ctx, "there is more data to sync from plaid, continuing", "iter", iter)
		}

		// Then enqueue all of the bank accounts we touched to have their similar
		// transactions recalculated.
		for key := range s.similarity {
			if err := queue.Enqueue(
				ctx,
				ctx.Enqueuer(),
				similar_jobs.CalculateTransactionClusters,
				s.similarity[key],
			); err != nil {
				return err
			}
		}

		return s.maintainLinkStatus(ctx, plaidLink)
	})
}
