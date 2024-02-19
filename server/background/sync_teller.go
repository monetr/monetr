package background

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/teller"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	SyncTeller = "SyncTeller"
)

var (
	_ Job = &SyncTellerJob{}
)

type (
	SyncTellerHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		kms          secrets.KeyManagement
		tellerClient teller.Client
		publisher    pubsub.Publisher
		enqueuer     JobEnqueuer
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	SyncTellerArguments struct {
		AccountId uint64 `json:"accountId"`
		LinkId    uint64 `json:"linkId"`
		// Trigger will be "visit", "cron", "manual", or "command"
		Trigger string `json:"trigger"`
	}

	SyncTellerJob struct {
		args      SyncTellerArguments
		log       *logrus.Entry
		repo      repository.BaseRepository
		secrets   repository.SecretsRepository
		teller    teller.Client
		publisher pubsub.Publisher
		enqueuer  JobEnqueuer
		clock     clock.Clock

		// Created by the job itself
		client            teller.AuthenticatedClient
		link              *models.Link
		timezone          *time.Location
		bankAccounts      map[string]models.BankAccount
		needsTransactions map[string]struct{}
		transactions      map[string]models.Transaction
		similarity        map[uint64]CalculateTransactionClustersArguments
		actions           map[uint64]SyncAction
		needsBalance      map[string]struct{}
		netChanges        map[string]int64

		tellerAccounts       map[string]teller.Account
		tellerTransactions   map[string]teller.Transaction
		tellerTransactionIds []string
	}
)

func TriggerSyncTeller(
	ctx context.Context,
	backgroundJobs JobController,
	arguments SyncTellerArguments,
) error {
	if arguments.Trigger == "" {
		arguments.Trigger = "manual"
	}
	return backgroundJobs.EnqueueJob(ctx, SyncTeller, arguments)
}

func NewSyncTellerHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	kms secrets.KeyManagement,
	tellerClient teller.Client,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
) *SyncTellerHandler {
	return &SyncTellerHandler{
		log:          log,
		db:           db,
		kms:          kms,
		tellerClient: tellerClient,
		publisher:    publisher,
		enqueuer:     enqueuer,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (s SyncTellerHandler) QueueName() string {
	return SyncTeller
}

func (s *SyncTellerHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args SyncTellerArguments
	if err := errors.Wrap(s.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Sync Teller job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return s.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := s.log.WithContext(span.Context())

		repo := repository.NewRepositoryFromSession(s.clock, 0, args.AccountId, txn)
		secretsRepo := repository.NewSecretsRepository(
			log,
			s.clock,
			txn,
			s.kms,
			args.AccountId,
		)
		job, err := NewSyncTellerJob(
			log,
			repo,
			s.clock,
			secretsRepo,
			s.tellerClient,
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

func (s SyncTellerHandler) DefaultSchedule() string {
	// Run every 12 hours. Links that have not received any updates in the last 13
	// hours will be synced with Teller. If no updates have been detected then
	// nothing will happen.
	return "0 0 */6 * * *"
}

func (s *SyncTellerHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := s.log.WithContext(ctx)

	log.Info("retrieving links to sync with Teller")

	links := make([]models.Link, 0)
	cutoff := s.clock.Now().Add(-3 * time.Hour)
	err := s.db.ModelContext(ctx, &links).
		Join(`INNER JOIN "teller_links" AS "teller_link"`).
		JoinOn(`"teller_link"."teller_link_id" = "link"."teller_link_id" AND "teller_link"."account_id" = "link"."account_id"`).
		Where(`"link"."link_type" = ?`, models.TellerLinkType).
		Where(`"teller_link"."status" = ?`, models.TellerLinkStatusSetup).
		Where(`"teller_link"."last_attempted_update" < ? OR "teller_link"."last_attempted_update" IS NULL`, cutoff).
		Where(`"link"."deleted_at" IS NULL`).
		Select(&links)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve links that need to by synced with Teller")
	}

	if len(links) == 0 {
		log.Debug("no Teller links need to be synced at this time")
		return nil
	}

	log.WithField("count", len(links)).Info("syncing Teller links")

	for _, item := range links {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
			"linkId":    item.LinkId,
		})
		itemLog.Trace("enqueuing link to be synced with Teller")
		err := enqueuer.EnqueueJob(ctx, s.QueueName(), SyncTellerArguments{
			AccountId: item.AccountId,
			LinkId:    item.LinkId,
			Trigger:   "cron",
		})
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job to sync with Teller")
			crumbs.Warn(ctx, "Failed to enqueue job to sync with Teller", "job", map[string]interface{}{
				"error": err,
			})
			continue
		}

		itemLog.Trace("successfully enqueued link to be synced with Teller")
	}

	return nil
}

func NewSyncTellerJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	secrets repository.SecretsRepository,
	tellerClient teller.Client,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
	args SyncTellerArguments,
) (*SyncTellerJob, error) {
	return &SyncTellerJob{
		args:      args,
		log:       log,
		repo:      repo,
		secrets:   secrets,
		teller:    tellerClient,
		publisher: publisher,
		enqueuer:  enqueuer,
		clock:     clock,

		client:             nil,
		link:               nil,
		timezone:           nil,
		bankAccounts:       map[string]models.BankAccount{},
		needsTransactions:  map[string]struct{}{},
		transactions:       map[string]models.Transaction{},
		similarity:         map[uint64]CalculateTransactionClustersArguments{},
		actions:            map[uint64]SyncAction{},
		needsBalance:       map[string]struct{}{},
		netChanges:         map[string]int64{},
		tellerAccounts:     map[string]teller.Account{},
		tellerTransactions: map[string]teller.Transaction{},
	}, nil
}

func (s *SyncTellerJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	s.log = s.log.WithContext(span.Context())

	link, err := s.repo.GetLink(span.Context(), s.args.LinkId)
	if err = errors.Wrap(err, "failed to retrieve link to sync with Teller"); err != nil {
		s.log.WithError(err).Error("cannot sync without link")
		return err
	}

	if link.TellerLink == nil {
		s.log.Warn("provided link does not have any Teller details")
		crumbs.IndicateBug(
			span.Context(),
			"BUG: Link was queued to sync with Teller, but has no Teller details",
			map[string]interface{}{
				"link": link,
			},
		)
		span.Status = sentry.SpanStatusFailedPrecondition
		return nil
	}
	s.link = link

	account, err := s.repo.GetAccount(span.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to retrieve account for job")
		return err
	}

	s.timezone, err = account.GetTimezone()
	if err != nil {
		s.log.WithError(err).Warn("failed to get account's time zone, defaulting to UTC")
		s.timezone = time.UTC
	}

	tellerLink := link.TellerLink
	s.log = s.log.WithFields(logrus.Fields{
		"tellerEnrollmentId": tellerLink.EnrollmentId,
		"tellerUserId":       tellerLink.UserId,
	})

	secret, err := s.secrets.Read(span.Context(), tellerLink.SecretId)
	if err != nil {
		s.log.WithError(err).Error("could not retrieve Teller secret for sync")
		return nil
	}

	s.client = s.teller.GetAuthenticatedClient(secret.Secret)

	// Before we do anything we need to sync the bank accounts that we will be
	// working with. This will setup any accounts that we don't already have
	// created as well as update the status of accounts that we have stored.
	if err := s.syncBankAccounts(span.Context()); err != nil {
		return err
	}

	// Sync the transactions once we know which accounts we want to sync.
	if err := s.syncTransactions(span.Context()); err != nil {
		return err
	}

	// Once we have synced all of the transactions we need to sync the account
	// balances. Some accounts will have their balance calculated by the net
	// transaction changes observed, and others will have their balance hard
	// queried from teller.
	if err := s.syncBalances(span.Context()); err != nil {
		return err
	}

	// Maintain the link status, when we perform the initial sync this will also
	// fire an event.
	if err := s.syncLinkStatus(span.Context()); err != nil {
		return err
	}

	// Then enqueue all of the bank accounts we touched to have their similar
	// transactions recalculated.
	for key := range s.similarity {
		s.enqueuer.EnqueueJob(span.Context(), CalculateTransactionClusters, s.similarity[key])
	}

	return nil
}

// syncBankAccounts will retrieve all of the current bank accounts that monetr
// has stored for the current link and the accounts from teller. It will create
// any accounts in monetr that don't already exist and update any existing
// accounts. For new accounts it will flag them as needing a balance update.
func (s *SyncTellerJob) syncBankAccounts(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	{ // Load the bank accounts that we already have into memory.
		bankAccounts, err := s.repo.GetBankAccountsByLinkId(span.Context(), s.args.LinkId)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve bank accounts for Teller sync")
		}
		for _, account := range bankAccounts {
			if account.TellerBankAccount == nil {
				s.log.WithField("bankAccountId", account.BankAccountId).
					Warn("bank account is part of a teller link, but does not have a teller bank account associated with it; it will be skipped")
				continue
			}
			s.bankAccounts[account.TellerBankAccount.TellerId] = account
		}
	}

	{ // Load the accounts from teller's API into memory.
		s.log.Debug("retrieving bank accounts from teller")
		tellerAccounts, err := s.client.GetAccounts(span.Context())
		if err != nil {
			s.log.WithError(err).Error("failed to retrieve bank accounts from teller")
			return err
		}

		if len(tellerAccounts) == 0 {
			s.log.Warn("no accounts found from Teller, something is wrong")
			return errors.New("no Teller accounts found")
		}

		s.log.Tracef("found %d account(s) from Teller", len(tellerAccounts))

		for _, account := range tellerAccounts {
			s.tellerAccounts[account.Id] = account
		}

		crumbs.AddTag(
			span.Context(),
			"teller.institution_id",
			tellerAccounts[0].Institution.Id,
		)
	}

	var err error
	for tellerId, account := range s.tellerAccounts {
		log := s.log.WithField("tellerAccountId", tellerId)
		log.Trace("syncing Teller account")
		bankAccount, ok := s.bankAccounts[tellerId]
		if !ok {
			log.Debug("Teller account has not been created in monetr, creating now")
			tellerBankAccount := models.TellerBankAccount{
				TellerLinkId:    *s.link.TellerLinkId,
				TellerId:        tellerId,
				InstitutionId:   account.Institution.Id,
				InstitutionName: account.Institution.Name,
				Mask:            account.Mask,
				Name:            account.Name,
				Type:            string(account.Type),
				SubType:         string(account.SubType),
				LedgerBalance:   nil, // Will be calculated or handled later
			}
			switch account.Status {
			case teller.AccountStatusClosed:
				tellerBankAccount.Status = models.TellerBankAccountStatusClosed
			case teller.AccountStatusOpen:
				tellerBankAccount.Status = models.TellerBankAccountStatusOpen
			default:
				panic("unrecognized teller account status")
			}
			if err = s.repo.CreateTellerBankAccount(
				span.Context(),
				&tellerBankAccount,
			); err != nil {
				return err
			}

			bankAccount = models.BankAccount{
				LinkId:              s.args.LinkId,
				TellerBankAccountId: &tellerBankAccount.TellerBankAccountId,
				TellerBankAccount:   &tellerBankAccount,
				AvailableBalance:    0,
				CurrentBalance:      0,
				Mask:                account.Mask,
				Name:                account.Name,
				OriginalName:        account.Name,
				Type:                s.getAccountType(account.Type),
				SubType:             s.getAccountSubType(account.SubType),
				Status:              s.getAccountStatus(account.Status),
			}
			if err = s.repo.CreateBankAccounts(
				span.Context(),
				&bankAccount,
			); err != nil {
				return err
			}

			log.WithFields(logrus.Fields{
				"tellerBankAccountId": tellerBankAccount.TellerBankAccountId,
				"bankAccountId":       bankAccount.BankAccountId,
			}).Debug("new bank account created from Teller, balance will be synced")

			s.bankAccounts[tellerId] = bankAccount
			s.flagNeedsBalance(tellerId)
			s.flagNeedsTransactions(tellerId)
			continue
		}

		log.Trace("a bank account for the Teller account already exists, checking for changes")
		changes := make([]SyncChange, 0)

		if bankAccount.Mask != account.Mask {
			changes = append(changes, SyncChange{
				Field: "mask",
				Old:   bankAccount.Mask,
				New:   account.Mask,
			})
			bankAccount.TellerBankAccount.Mask = account.Mask
			bankAccount.Mask = account.Mask
		}

		if bankAccount.OriginalName != account.Name {
			changes = append(changes, SyncChange{
				Field: "originalName",
				Old:   bankAccount.OriginalName,
				New:   account.Name,
			})
			bankAccount.TellerBankAccount.Name = account.Name
			bankAccount.OriginalName = account.Name
		}

		if status := s.getAccountStatus(account.Status); bankAccount.Status != status {
			changes = append(changes, SyncChange{
				Field: "status",
				Old:   bankAccount.Status,
				New:   status,
			})
			// TODO Update the teller bank account status too!
			bankAccount.Status = status
			// If the status of an account changes (regardless of what it changes to)
			// then we want to pull transactions for that account. This might be from
			// an active to an inactive status. In which case this will be the last
			// time we retrieve transactions for that account.
			s.flagNeedsTransactions(tellerId)
		} else if bankAccount.Status == models.ActiveBankAccountStatus {
			// If an account is active though, we should just always retrieve it's
			// transactions.
			s.flagNeedsTransactions(tellerId)
		}

		if len(changes) == 0 {
			log.Debug("no changes derected for bank account")
			continue
		}

		if err = s.repo.UpdateTellerBankAccount(
			span.Context(),
			bankAccount.TellerBankAccount,
		); err != nil {
			return err
		}
		if err = s.repo.UpdateBankAccounts(span.Context(), bankAccount); err != nil {
			return err
		}

		// Update cached data
		s.bankAccounts[tellerId] = bankAccount
	}

	return nil
}

func (s *SyncTellerJob) retrieveTellerTransactions(
	ctx context.Context,
	log *logrus.Entry,
	tellerId string,
	immutableTimestamp *time.Time,
) error {
	tellerTransactions := make([]teller.Transaction, 0)
	var pageSize int64 = 25
	for {
		var fromId *string
		if length := len(tellerTransactions); length > 0 {
			fromId = &tellerTransactions[length-1].Id
		}
		transactions, err := s.client.GetTransactions(
			ctx,
			tellerId,
			fromId,
			pageSize,
		)
		if err != nil {
			log.WithField("fromId", fromId).
				WithError(err).
				Error("failed to retrieve transactions from Teller")
			return err
		}

		tellerTransactions = append(tellerTransactions, transactions...)
		if len(transactions) < int(pageSize) {
			// If we receive fewer than the number we requested that means we have
			// reached the end of the list.
			break
		}

		// If we do not have an immutable timestamp then keep requesting until we
		// run out of transactions.
		if immutableTimestamp == nil {
			continue
		}

		// If we do have an immutable timestamp though then only request
		// transactions until we find one older than the date we are working with.
		if length := len(tellerTransactions); length > 0 {
			last := tellerTransactions[length-1]
			date, err := last.GetDate(s.timezone)
			if err != nil {
				return err
			}
			// Only if the date is before do we stop, if the date is the same that's
			// fine.
			if date.Before(*immutableTimestamp) {
				break
			}
		}
	}

	// Clear out the transactions from a previous account. We are only working
	// with a single account at a time.
	s.tellerTransactions = make(map[string]teller.Transaction)
	s.tellerTransactionIds = make([]string, 0, len(tellerTransactions))

	// Cache the transactions we have retrieve first
	for _, tellerTransaction := range tellerTransactions {
		// Throw out transactions who are older than our immutable timestamp.
		date, err := tellerTransaction.GetDate(s.timezone)
		if err != nil {
			return err
		}

		// If we actually have an immutable timestamp for this sync.
		if immutableTimestamp != nil {
			// Then throw out any transactions that we did retrieve that are before
			// that timestamp.
			if date.Before(*immutableTimestamp) {
				continue
			}
		}

		s.tellerTransactions[tellerTransaction.Id] = tellerTransaction
		s.tellerTransactionIds = append(s.tellerTransactionIds, tellerTransaction.Id)
	}

	return nil
}

func (s *SyncTellerJob) getNewImmutableTimestamp() (time.Time, error) {
	// Now calculate the new immutable timestamp based on the transactions we
	// just received.
	var newImmutableTimestamp time.Time
	// Find the oldest pending transaction and use that
	for _, txn := range s.tellerTransactions {
		if txn.Status == teller.TransactionStatusPending {
			date, err := txn.GetDate(s.timezone)
			if err != nil {
				return newImmutableTimestamp, err
			}
			date = date.AddDate(0, 0, -1)
			if newImmutableTimestamp.IsZero() || date.Before(newImmutableTimestamp) {
				newImmutableTimestamp = date
			}
		}
	}
	// If there wasn't one then use the latest transaction's date.
	if newImmutableTimestamp.IsZero() {
		for _, txn := range s.tellerTransactions {
			date, err := txn.GetDate(s.timezone)
			if err != nil {
				return newImmutableTimestamp, err
			}
			date = date.AddDate(0, 0, -1)
			if newImmutableTimestamp.IsZero() || date.Before(newImmutableTimestamp) {
				newImmutableTimestamp = date
			}
		}
	}

	if newImmutableTimestamp.IsZero() {
		newImmutableTimestamp = s.clock.Now().AddDate(0, 0, -7)
	}

	return newImmutableTimestamp, nil
}

func (s *SyncTellerJob) syncTransactions(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	if len(s.needsTransactions) == 0 {
		s.log.Info("no accounts need their transactions synced at this time")
		return nil
	}

	for tellerId := range s.needsTransactions {
		bankAccount := s.bankAccounts[tellerId]
		log := s.log.WithFields(logrus.Fields{
			"tellerAccountId": tellerId,
			"bankAccountId":   bankAccount.BankAccountId,
		})

		latestSync, err := s.repo.GetLatestTellerSync(span.Context(), *bankAccount.TellerBankAccountId)
		if err != nil {
			log.WithError(err).Error("failed to determine latest sync data, will skip syncing transactions for account")
			continue
		}

		var immutableTimestamp *time.Time
		if latestSync != nil {
			immutableTimestamp = &latestSync.ImmutableTimestamp
		}

		// Retrieve the transactions from teller and store them first.
		if err := s.retrieveTellerTransactions(
			span.Context(),
			log,
			tellerId,
			immutableTimestamp,
		); err != nil {
			log.WithError(err).Error("failed to retrieve transactions from teller")
			continue
		}

		{ // Retrieve stored transaction data to compare to the API results from teller.
			workingTransactions, err := s.repo.GetTransactionsByTellerId(
				span.Context(),
				bankAccount.BankAccountId,
				s.tellerTransactionIds,
				true, // Include pending transactions
			)
			if err != nil {
				log.WithError(err).Error("failed to read working transactions from the database")
				continue
			}

			s.transactions = make(map[string]models.Transaction)
			for _, transaction := range workingTransactions {
				if transaction.TellerTransaction == nil {
					continue
				}
				s.transactions[transaction.TellerTransaction.TellerId] = transaction
			}
		}

		for tellerTransactionId, tellerTxnRaw := range s.tellerTransactions {
			txnLog := log.WithFields(logrus.Fields{
				"tellerTransactionId": tellerTransactionId,
			})

			amount, err := tellerTxnRaw.GetAmount()
			if err != nil {
				return err
			}

			runningBalance, err := tellerTxnRaw.GetRunningBalance()
			if err != nil {
				return err
			}

			date, err := tellerTxnRaw.GetDate(s.timezone)
			if err != nil {
				return err
			}

			isPending := tellerTxnRaw.Status == teller.TransactionStatusPending

			transaction, ok := s.transactions[tellerTransactionId]
			// If the transaction does not exist
			if !ok {
				txnLog.Trace("transaction does not exist, it will be created")
				tellerTransaction := models.TellerTransaction{
					TellerBankAccountId: *bankAccount.TellerBankAccountId,
					TellerId:            tellerTransactionId,
					Name:                tellerTxnRaw.GetDescription(),
					Category:            string(tellerTxnRaw.Details.Category),
					Type:                tellerTxnRaw.Type,
					Date:                date,
					IsPending:           isPending,
					Amount:              amount,
					RunningBalance:      runningBalance,
				}
				if err := s.repo.CreateTellerTransaction(
					span.Context(),
					&tellerTransaction,
				); err != nil {
					crumbs.Error(span.Context(), "Failed to insert teller transaction", "teller", map[string]interface{}{
						"tellerTransactionId": tellerTransactionId,
						"tellerAccountId":     tellerId,
						"bankAccountId":       bankAccount.BankAccountId,
						"immutableTimestamp":  immutableTimestamp,
						"date":                date,
					})
					return err
				}

				transaction = models.Transaction{
					BankAccountId:        bankAccount.BankAccountId,
					TellerTransactionId:  &tellerTransaction.TellerTransactionId,
					TellerTransaction:    &tellerTransaction,
					Amount:               amount,
					Categories:           nil,
					Date:                 date,
					Name:                 tellerTxnRaw.GetDescription(),
					OriginalName:         tellerTxnRaw.GetDescription(),
					MerchantName:         tellerTxnRaw.Details.Counterparty.Name,
					OriginalMerchantName: tellerTxnRaw.Details.Counterparty.Name,
					Currency:             "USD", // TODO Derive this from somewhere
					IsPending:            isPending,
				}
				if err := s.repo.CreateTransaction(
					span.Context(),
					bankAccount.BankAccountId,
					&transaction,
				); err != nil {
					return err
				}
				// Subtract the transaction balance from this account's balance but only
				// if the transaction is not pending.
				if !isPending {
					s.netChanges[tellerId] -= amount
				} else {
					// Flag the account for updating, but don't affect main balance. Net changes only affects the
					// current balance, but by flagging it like this we make sure that it will be recalculated for the
					// available balance based on the pending transactions.
					s.netChanges[tellerId] += 0
				}
				s.tagBankAccountForSimilarityRecalc(bankAccount.BankAccountId)
				continue
			}

			txnLog = txnLog.WithField("transactionId", transaction.TransactionId)

			changes := make([]SyncChange, 0)

			if transaction.Amount != amount {
				changes = append(changes, SyncChange{
					Field: "amount",
					Old:   transaction.Amount,
					New:   amount,
				})
				// When the amount of a transaction changes we need to adjust the net
				// for this account by the delta of that change. But we have to add it
				// rather than subtract it because its the delta.
				delta := transaction.Amount - amount
				s.netChanges[tellerId] += delta
				transaction.Amount = amount
			}

			if transaction.IsPending != isPending {
				changes = append(changes, SyncChange{
					Field: "isPending",
					Old:   transaction.IsPending,
					New:   isPending,
				})
				// When we transition to a posted status then we need to affect the
				// balance.
				if !isPending {
					s.netChanges[tellerId] -= transaction.Amount
				} else {
					s.netChanges[tellerId] += 0
				}
				transaction.IsPending = isPending
			}

			if len(changes) == 0 {
				txnLog.Trace("no changes detected for transaction, nothing to be done")
				continue
			}

			if err := s.repo.UpdateTransaction(span.Context(), bankAccount.BankAccountId, &transaction); err != nil {
				return err
			}
			s.tagBankAccountForSimilarityRecalc(bankAccount.BankAccountId)
		}

		// Check for and remove any transactions that are no longer present in the
		// list from teller.
		if err := s.syncRemovedTransactions(span.Context(), log, tellerId); err != nil {
			return err
		}

		newImmutableTimestamp, err := s.getNewImmutableTimestamp()
		if err != nil {
			return err
		}

		// TODO It might be nice to add the balance sync boi here instead. That way
		// we can log the balance with the sync. Or lift this to be afterwards?
		if err := s.repo.CreateTellerSync(span.Context(), &models.TellerSync{
			TellerBankAccountId: *bankAccount.TellerBankAccountId,
			Timestamp:           s.clock.Now(),
			Trigger:             s.args.Trigger,
			ImmutableTimestamp:  newImmutableTimestamp,
			Added:               0, // TODO
			Modified:            0,
			Removed:             0,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *SyncTellerJob) syncRemovedTransactions(
	ctx context.Context,
	log *logrus.Entry,
	tellerId string,
) error {
	// Detect deleted transactions
	for tellerTransactionId, transaction := range s.transactions {
		_, ok := s.tellerTransactions[tellerTransactionId]
		if ok {
			continue
		}

		if !transaction.IsPending {
			crumbs.IndicateBug(ctx, "Trying to remove a posted transaction, this is a bug", map[string]interface{}{
				"tellerAccountId":     tellerId,
				"tellerTransactionId": tellerTransactionId,
				"transactionId":       transaction.TransactionId,
				"bankAccountId":       transaction.BankAccountId,
			})
		}

		existing := transaction

		txnLog := log.WithFields(logrus.Fields{
			"tellerTransactionId": tellerTransactionId,
			"transactionId":       transaction.TransactionId,
		})
		txnLog.Debug("transaction is not deleted in monetr but is missing in teller, it will be removed")

		// Add the amount of the transaction back to the net balance.
		if !transaction.IsPending {
			s.netChanges[tellerId] += transaction.Amount
		} else {
			s.netChanges[tellerId] += 0
		}

		// If the transaction has a spending object associated with it, remove it.
		if transaction.SpendingId != nil {
			updated := transaction
			// Unset the spending ID on the updated one, but not the existing one.
			updated.SpendingId = nil
			if existing.SpendingId == nil {
				panic("WHY DOES MEMORY WORK LIKE THIS")
			}

			transaction.SpendingId = nil
			_, err := s.repo.ProcessTransactionSpentFrom(
				ctx,
				transaction.BankAccountId,
				&updated,
				&existing,
			)
			if err != nil {
				return err
			}
		}

		if err := s.repo.DeleteTransaction(
			ctx,
			transaction.BankAccountId,
			transaction.TransactionId,
		); err != nil {
			return err
		}
		if err := s.repo.DeleteTellerTransaction(
			ctx,
			transaction.TellerTransaction,
		); err != nil {
			return err
		}
		s.tagBankAccountForSimilarityRecalc(transaction.BankAccountId)
	}

	return nil
}

func (s *SyncTellerJob) syncBalances(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	net := make([]string, 0)
	for tellerId := range s.needsBalance {
		bankAccount := s.bankAccounts[tellerId]
		log := s.log.WithFields(logrus.Fields{
			"tellerAccountId": tellerId,
			"bankAccountId":   bankAccount.BankAccountId,
			"balancedAt":      bankAccount.TellerBankAccount.BalancedAt,
		})

		if !s.canQueryBalance(bankAccount) {
			log.Warn("account balance was hard queried too recently, it will be skipped")
			net = append(net, tellerId)
			continue
		}

		log.Info("hard querying account balance from Teller")
		balance, err := s.client.GetAccountBalance(span.Context(), tellerId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve account balance from Teller")
			return errors.Wrap(err, "failed to hard sync balance for account")
		}

		currentBalance := bankAccount.CurrentBalance
		currentAvailable := bankAccount.AvailableBalance

		pendingBalance, err := s.getPendingTransactionBalance(
			span.Context(),
			bankAccount.BankAccountId,
		)
		if err != nil {
			return err
		}

		bankAccount.CurrentBalance, err = balance.GetLedger()
		if err != nil {
			return err
		}

		bankAccount.AvailableBalance = bankAccount.CurrentBalance - pendingBalance

		tellerAvailable, err := balance.GetAvailable()
		if err != nil {
			return err
		}

		log.WithFields(logrus.Fields{
			"oldCurrent":      currentBalance,
			"newCurrent":      bankAccount.CurrentBalance,
			"oldAvailable":    currentAvailable,
			"newAvailable":    bankAccount.AvailableBalance,
			"pendingBalance":  pendingBalance,
			"tellerAvailable": tellerAvailable,
		}).Debug("updating bank account balance")

		if err := s.repo.UpdateBankAccounts(span.Context(), bankAccount); err != nil {
			return err
		}

		// Update cache with the new balances incase we use it again.
		s.bankAccounts[tellerId] = bankAccount
	}

	// If we were not able to update the balance for some accounts then try to
	// update them via netting instead.
	for _, tellerId := range net {
		delete(s.needsBalance, tellerId)
	}

	for tellerId, net := range s.netChanges {
		bankAccount, ok := s.bankAccounts[tellerId]
		if !ok {
			s.log.
				WithField("tellerAccountId", tellerId).
				Warn("could not sync net changes for bank account, something is wrong")
			continue
		}

		log := s.log.WithFields(logrus.Fields{
			"tellerAccountId": tellerId,
			"bankAccountId":   bankAccount.BankAccountId,
		})

		if _, ok := s.needsBalance[tellerId]; ok {
			log.Info("bank account balance will be updated via Teller API instead of net changes")
			continue
		}

		currentBalance := bankAccount.CurrentBalance
		currentAvailable := bankAccount.AvailableBalance
		bankAccount.CurrentBalance += net
		pendingBalance, err := s.getPendingTransactionBalance(
			span.Context(),
			bankAccount.BankAccountId,
		)
		if err != nil {
			return err
		}

		bankAccount.AvailableBalance = bankAccount.CurrentBalance - pendingBalance

		log.WithFields(logrus.Fields{
			"oldCurrent":     currentBalance,
			"newCurrent":     bankAccount.CurrentBalance,
			"oldAvailable":   currentAvailable,
			"newAvailable":   bankAccount.AvailableBalance,
			"pendingBalance": pendingBalance,
			"netBalance":     net,
		}).Debug("updating bank account balance")

		if err := s.repo.UpdateBankAccounts(span.Context(), bankAccount); err != nil {
			return err
		}

		// Update cache with the new balances incase we use it again.
		s.bankAccounts[tellerId] = bankAccount
	}

	return nil
}

func (s *SyncTellerJob) syncLinkStatus(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	tellerLink := s.link.TellerLink

	linkWasSetup := false
	switch tellerLink.Status {
	case models.TellerLinkStatusPending, models.TellerLinkStatusUnknown:
		crumbs.Debug(ctx, "Updating Teller link status.", map[string]interface{}{
			"old": tellerLink.Status,
			"new": models.TellerLinkStatusSetup,
		})
		tellerLink.Status = models.TellerLinkStatusSetup
		linkWasSetup = true
	}
	tellerLink.LastSuccessfulUpdate = myownsanity.TimeP(s.clock.Now().UTC())
	tellerLink.LastAttemptedUpdate = myownsanity.TimeP(s.clock.Now().UTC())
	if err := s.repo.UpdateTellerLink(span.Context(), tellerLink); err != nil {
		return err
	}

	if linkWasSetup {
		channelName := fmt.Sprintf("initial:teller:link:%d:%d", s.args.AccountId, s.args.LinkId)
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

func (s *SyncTellerJob) getPendingTransactionBalance(ctx context.Context, bankAccountId uint64) (int64, error) {
	var pendingBalance int64
	offset := 0
	for { // Retrieve all of the pending transactions for the current account.
		pageSize := 10
		pendingTransactions, err := s.repo.GetPendingTransactions(
			ctx,
			bankAccountId,
			// Limit and offset, keep these small so we make a few more queries but
			// we don't force the DB to overread a ton. If we set the limit high
			// here and we don't have any pending transactions then the DB will work
			// a LOT harder.
			// TODO We could also limit by date here, like don't consider pending
			// transactions older than 30 days?
			pageSize, offset,
		)
		if err != nil {
			return 0, errors.Wrap(err, "failed to read pending transactions to calculate available balance")
		}

		// Rollup the sum of the pending transactions
		for _, transaction := range pendingTransactions {
			pendingBalance += transaction.Amount
		}
		offset += len(pendingTransactions)

		// If we have fewer items in our result than we requested for the page
		// size then that means there aren't anymore pending transactions.
		if len(pendingTransactions) < pageSize {
			break
		}
	}

	return pendingBalance, nil
}

// canQueryBalance takes a bank account with a teller bank account sub object,
// it will return true if the account can have it's balance hard queried from
// teller. I don't want to allow accounts to have their balance queried
// frequently since we pay per API call for that endpoint. Instead we should
// make sure we can only do a hard query once every 30 days. So if we have made
// a hard balance update in the past 30 days, return false and do not allow an
// update.
func (s *SyncTellerJob) canQueryBalance(bankAccount models.BankAccount) bool {
	balancedAt := bankAccount.TellerBankAccount.BalancedAt
	if balancedAt == nil {
		return true
	}

	if balancedAt.Add(30 * 24 * time.Hour).Before(s.clock.Now()) {
		return true
	}

	return false
}

func (s *SyncTellerJob) getAccountType(tellerType teller.AccountType) models.BankAccountType {
	switch tellerType {
	case teller.AccountTypeCredit:
		return models.CreditBankAccountType
	case teller.AccountTypeDepository:
		return models.DepositoryBankAccountType
	default:
		return models.OtherBankAccountType
	}
}

func (s *SyncTellerJob) getAccountSubType(tellerType teller.AccountSubType) models.BankAccountSubType {
	switch tellerType {
	case teller.AccountSubTypeChecking:
		return models.CheckingBankAccountSubType
	case teller.AccountSubTypeSavings:
		return models.SavingsBankAccountSubType
	case teller.AccountSubTypeMoneyMarket:
		return models.MoneyMarketBankAccountSubType
	case teller.AccountSubTypeCertificateOfDeposit:
		return models.CDBankAccountSubType
	case teller.AccountSubTypeCreditCard:
		return models.CreditCardBankAccountSubType
	case teller.AccountSubTypeTreasury, teller.AccountSubTypeSweep:
		fallthrough
	default:
		return models.OtherBankAccountSubType
	}

}

func (s *SyncTellerJob) getAccountStatus(tellerStatus teller.AccountStatus) models.BankAccountStatus {
	switch tellerStatus {
	case teller.AccountStatusClosed:
		return models.InactiveBankAccountStatus
	case teller.AccountStatusOpen:
		return models.ActiveBankAccountStatus
	default:
		return models.UnknownBankAccountStatus
	}
}

// flagNeedsBalance will mark a bank account as needing to have the balance hard
// queried from Teller. This is because it is not clear when teller will include
// the running balance on the transactions, so monetr needs to keep track of the
// balance on its own. However certain events like an account being created
// require us to hard query the balance to get a base number.
func (s *SyncTellerJob) flagNeedsBalance(tellerBankAccountId string) {
	s.needsBalance[tellerBankAccountId] = struct{}{}
}

// flagNeedsTransactions will mark a bank account as needing to have it's
// transactions retrieved from the API. This is so that we don't retrieve
// transactions for every account every time. If an account is closed for
// example, we don't need to retrieve it's transactions anymore.
func (s *SyncTellerJob) flagNeedsTransactions(tellerBankAccountId string) {
	s.needsTransactions[tellerBankAccountId] = struct{}{}
}

func (s *SyncTellerJob) tagBankAccountForSimilarityRecalc(bankAccountId uint64) {
	s.similarity[bankAccountId] = CalculateTransactionClustersArguments{
		AccountId:     s.args.AccountId,
		BankAccountId: bankAccountId,
	}
}
