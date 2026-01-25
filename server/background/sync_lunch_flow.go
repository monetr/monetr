package background

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	SyncLunchFlow = "SyncLunchFlow"
)

type LunchFlowSyncStatus string

const (
	LunchFlowSyncStatusBegin        LunchFlowSyncStatus = "begin"
	LunchFlowSyncStatusTransactions LunchFlowSyncStatus = "transactions"
	LunchFlowSyncStatusBalances     LunchFlowSyncStatus = "balances"
	LunchFlowSyncStatusComplete     LunchFlowSyncStatus = "complete"
	LunchFlowSyncStatusError        LunchFlowSyncStatus = "error"
)

var (
	_ ScheduledJobHandler = &SyncLunchFlowHandler{}
	_ JobImplementation   = &SyncLunchFlowJob{}
)

type (
	SyncLunchFlowHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		kms          secrets.KeyManagement
		publisher    pubsub.Publisher
		enqueuer     JobEnqueuer
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	SyncLunchFlowArguments struct {
		AccountId     ID[Account]     `json:"accountId"`
		BankAccountId ID[BankAccount] `json:"bankAccountId"`
		LinkId        ID[Link]        `json:"linkId"`
	}

	SyncLunchFlowJob struct {
		args      SyncLunchFlowArguments
		log       *logrus.Entry
		repo      repository.BaseRepository
		secrets   repository.SecretsRepository
		publisher pubsub.Publisher
		enqueuer  JobEnqueuer
		clock     clock.Clock
		timezone  *time.Location

		client                lunch_flow.LunchFlowClient
		bankAccount           *BankAccount
		lunchFlowTransactions []lunch_flow.Transaction
		existingTransactions  map[string]Transaction
	}
)

func TriggerSyncLunchFlowTxn(
	ctx context.Context,
	backgroundJobs JobController,
	txn pg.DBI,
	arguments SyncLunchFlowArguments,
) error {
	return backgroundJobs.EnqueueJobTxn(ctx, txn, SyncLunchFlow, arguments)
}

func NewSyncLunchFlowHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	kms secrets.KeyManagement,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
) *SyncLunchFlowHandler {
	return &SyncLunchFlowHandler{
		log:          log,
		db:           db,
		kms:          kms,
		publisher:    publisher,
		enqueuer:     enqueuer,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

// DefaultSchedule implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) DefaultSchedule() string {
	// Run every 12 hours at 30 minutes after the hour.
	return "0 30 */12 * * *"
}

// EnqueueTriggeredJob implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) EnqueueTriggeredJob(
	ctx context.Context,
	enqueuer JobEnqueuer,
) error {
	log := s.log.WithContext(ctx)

	log.Info("retrieving bank accounts to sync with Lunch Flow")

	jobRepo := repository.NewJobRepository(s.db, s.clock)

	bankAccounts, err := jobRepo.GetLunchFlowAccountsToSync(ctx)
	if err != nil {
		return err
	}

	if len(bankAccounts) == 0 {
		log.Info("no bank accounts to by synced with Lunch Flow at this time")
		return nil
	}

	for _, item := range bankAccounts {
		itemLog := log.WithFields(logrus.Fields{
			"accountId":     item.AccountId,
			"bankAccountId": item.BankAccountId,
			"linkId":        item.LinkId,
		})

		itemLog.Trace("enqueuing bank account to be synced with Lunch Flow")

		err := enqueuer.EnqueueJobTxn(
			ctx,
			s.db,
			s.QueueName(),
			SyncLunchFlowArguments{
				AccountId:     item.AccountId,
				BankAccountId: item.BankAccountId,
				LinkId:        item.LinkId,
			},
		)
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job to sync with Lunch Flow")
			crumbs.Warn(ctx, "Failed to enqueue job to sync with Lunch Flow", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Trace("successfully enqueued bank account to be synced with Lunch Flow")
	}

	return nil
}

// HandleConsumeJob implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) HandleConsumeJob(ctx context.Context, log *logrus.Entry, data []byte) error {
	var args SyncLunchFlowArguments
	if err := errors.Wrap(s.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Sync Lunch Flow job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)
	log = log.WithFields(logrus.Fields{
		"accountId":     args.AccountId,
		"bankAccountId": args.BankAccountId,
	})

	return s.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context())
		repo := repository.NewRepositoryFromSession(
			s.clock,
			"user_lunch_flow",
			args.AccountId,
			txn,
			log,
		)
		secretsRepo := repository.NewSecretsRepository(
			log,
			s.clock,
			txn,
			s.kms,
			args.AccountId,
		)

		job, err := NewSyncLunchFlowJob(
			log,
			repo,
			s.clock,
			secretsRepo,
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

// QueueName implements ScheduledJobHandler.
func (s *SyncLunchFlowHandler) QueueName() string {
	return SyncLunchFlow
}

func NewSyncLunchFlowJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	secrets repository.SecretsRepository,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
	args SyncLunchFlowArguments,
) (*SyncLunchFlowJob, error) {
	return &SyncLunchFlowJob{
		args:      args,
		log:       log,
		repo:      repo,
		secrets:   secrets,
		publisher: publisher,
		enqueuer:  enqueuer,
		clock:     clock,
		timezone:  nil,

		existingTransactions: make(map[string]Transaction),
	}, nil
}

// Run implements JobImplementation.
func (s *SyncLunchFlowJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	s.progress(span.Context(), LunchFlowSyncStatusBegin)

	s.log = s.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"accountId":     s.args.AccountId,
		"bankAccountId": s.args.BankAccountId,
	})

	{ // Retrieve the bank account from the database, store the details for later
		bankAccount, err := s.repo.GetBankAccount(
			span.Context(),
			s.args.BankAccountId,
		)
		if err != nil {
			return err
		}

		if bankAccount.LunchFlowBankAccount == nil {
			s.log.Warn("bank account does not have a Lunch Flow account associated with it")
			span.Status = sentry.SpanStatusFailedPrecondition
			return nil
		}
		s.log = s.log.WithFields(logrus.Fields{
			"lunchFlow": logrus.Fields{
				"lunchFlowId":     bankAccount.LunchFlowBankAccount.LunchFlowId,
				"institutionName": bankAccount.LunchFlowBankAccount.InstitutionName,
			},
		})
		s.bankAccount = bankAccount
	}

	if err := s.setupClient(span.Context()); err != nil {
		return err
	}

	s.progress(span.Context(), LunchFlowSyncStatusTransactions)
	if err := s.hydrateTransactions(span.Context()); err != nil {
		return err
	}

	if err := s.syncTransactions(span.Context()); err != nil {
		return err
	}

	s.progress(span.Context(), LunchFlowSyncStatusBalances)
	if err := s.syncBalances(span.Context()); err != nil {
		return err
	}

	s.progress(span.Context(), LunchFlowSyncStatusComplete)
	s.log.Info("finished syncing Lunch Flow account")

	return nil
}

func (s *SyncLunchFlowJob) progress(ctx context.Context, status LunchFlowSyncStatus) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := s.log.WithContext(span.Context())

	log.WithField("status", status).Debug("sending progress update")

	channel := fmt.Sprintf(
		"account:%s:link:%s:bank_account:%s:lunch_flow_sync_progress",
		s.args.AccountId, s.args.LinkId, s.args.BankAccountId,
	)
	j, err := json.Marshal(map[string]any{
		"bankAccountId": s.args.BankAccountId,
		"status":        status,
	})
	if err != nil {
		log.WithError(err).Error("failed to encode progress message")
		return nil
	}
	if err := s.publisher.Notify(ctx, channel, string(j)); err != nil {
		return errors.Wrap(err, "failed to send progress notification for job")
	}

	return nil
}

func (s *SyncLunchFlowJob) setupClient(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	link, err := s.repo.GetLink(span.Context(), s.bankAccount.LinkId)
	if err != nil {
		return err
	}

	if link.LunchFlowLink == nil {
		s.log.Warn("provided link does not have any Lunch Flow credentials")
		crumbs.IndicateBug(
			span.Context(),
			"BUG: Link was queued to sync with Lunch Flow, but has no Lunch Flow details",
			map[string]any{
				"link": link,
			},
		)
		return nil
	}
	s.log = s.log.WithFields(logrus.Fields{
		"linkId":          link.LinkId,
		"lunchFlowLinkId": link.LunchFlowLinkId,
	})

	{ // Store the account's timezone information
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
	}

	{ // Retrieve the secret and setup the API client
		secret, err := s.secrets.Read(span.Context(), link.LunchFlowLink.SecretId)
		if err = errors.Wrap(err, "failed to retrieve credentials for Lunch Flow"); err != nil {
			s.log.WithError(err).Error("could not retrieve API credentials for Lunch Flow for link")
			return err
		}

		client, err := lunch_flow.NewLunchFlowClient(
			s.log,
			link.LunchFlowLink.ApiUrl,
			secret.Value,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create Lunch Flow API client")
		}

		s.client = client
	}

	return nil
}

func (s *SyncLunchFlowJob) hydrateTransactions(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.WithContext(span.Context()).Debug("fetching Lunch Flow transactions")

	var err error
	s.lunchFlowTransactions, err = s.client.GetTransactions(
		span.Context(),
		lunch_flow.AccountId(s.bankAccount.LunchFlowBankAccount.LunchFlowId),
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve transactions from Lunch Flow for sync")
	}

	lunchFlowIds := make([]string, len(s.lunchFlowTransactions))
	for i := range s.lunchFlowTransactions {
		lunchFlowIds[i] = s.lunchFlowTransactions[i].Id
	}

	s.existingTransactions, err = s.repo.GetTransactionsByLunchFlowId(
		span.Context(),
		s.bankAccount.BankAccountId,
		lunchFlowIds,
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve existing Lunch Flow transactions")
	}

	if count := len(s.existingTransactions); count > 0 {
		s.log.WithContext(span.Context()).WithFields(logrus.Fields{
			"existingTransactions": count,
		}).Debug("found existing transactions for sync")
	}

	return nil
}

func (s *SyncLunchFlowJob) syncTransactions(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := s.log.WithContext(span.Context())
	log.Debug("syncing Lunch Flow transactions")

	// transactionsToUpdate := make([]*Transaction, 0)
	lunchFlowTransactionsToCreate := make([]LunchFlowTransaction, 0)
	transactionsToCreate := make([]Transaction, 0)

	for i := range s.lunchFlowTransactions {
		externalTransaction := s.lunchFlowTransactions[i]
		tlog := log.WithFields(logrus.Fields{
			"lunchFlowTransaction": logrus.Fields{
				"lunchFlowId": externalTransaction.Id,
				"amount":      externalTransaction.Amount.String(),
				"currency":    externalTransaction.Currency,
			},
		})

		if externalTransaction.Currency != s.bankAccount.Currency {
			tlog.WithFields(logrus.Fields{
				"currency": logrus.Fields{
					"expected": s.bankAccount.Currency,
					"received": externalTransaction.Currency,
				},
			}).Warn("currency on transaction does not match bank account's stored currency, this may cause problems!")
		}

		amount, err := currency.ParseFriendlyToAmount(
			externalTransaction.Amount.String(),
			s.bankAccount.Currency,
		)
		if err != nil {
			tlog.WithError(err).Error("failed to parse transaction amount")
			continue
		}

		fallbackName := fmt.Sprintf(
			"Transaction %s %s %s",
			externalTransaction.Date,
			externalTransaction.Amount.String(),
			externalTransaction.Currency,
		)

		// Note the order of the fields here, for the name field we want to
		// prioritize the merchant name if there is one. If their isn't we'll fall
		// back to the description, and if there is nothing else then we will use
		// our generated fallback name.
		name := myownsanity.CoalesceStrings(
			externalTransaction.Merchant,
			externalTransaction.Description,
			fallbackName,
		)
		// Original name prioritizes the description instead.
		originalName := myownsanity.CoalesceStrings(
			externalTransaction.Description,
			externalTransaction.Merchant,
			fallbackName,
		)

		date, err := time.ParseInLocation(
			"2006-01-02",
			externalTransaction.Date,
			s.timezone,
		)
		if err != nil {
			log.
				WithError(err).
				WithFields(logrus.Fields{
					"value": externalTransaction.Date,
				}).
				Error("failed to parse transaction date from Lunch Flow")
			continue
		}

		transaction, ok := s.existingTransactions[externalTransaction.Id]
		if !ok {
			lunchFlowTransaction := LunchFlowTransaction{
				LunchFlowTransactionId: NewID(&LunchFlowTransaction{}),
				AccountId:              s.bankAccount.AccountId,
				LunchFlowBankAccountId: *s.bankAccount.LunchFlowBankAccountId,
				LunchFlowId:            externalTransaction.Id,
				Merchant:               externalTransaction.Merchant,
				Description:            externalTransaction.Description,
				Date:                   date,
				Currency:               externalTransaction.Currency,
				Amount:                 amount,
				IsPending:              false,
			}

			transaction = Transaction{
				AccountId:              s.bankAccount.AccountId,
				BankAccountId:          s.bankAccount.BankAccountId,
				LunchFlowTransactionId: &lunchFlowTransaction.LunchFlowTransactionId,
				LunchFlowTransaction:   &lunchFlowTransaction,
				Amount:                 amount,
				Date:                   date,
				Name:                   name,
				OriginalName:           originalName,
				MerchantName:           externalTransaction.Merchant,
				OriginalMerchantName:   externalTransaction.Merchant,
				IsPending:              false,
				Source:                 "lunch_flow",
			}
			lunchFlowTransactionsToCreate = append(lunchFlowTransactionsToCreate, lunchFlowTransaction)
			transactionsToCreate = append(transactionsToCreate, transaction)
			continue
		}
		// TODO Handle updating transactions too!
		tlog.Trace("transaction from Lunch Flow already exists in monetr, skipping")
		continue
	}

	// Persist any new transactions.
	if count := len(transactionsToCreate); count > 0 {
		log.WithField("new", count).Info("creating new Lunch Flow transactions")

		// Create Lunch Flow transactions first
		if err := s.repo.CreateLunchFlowTransactions(
			span.Context(),
			lunchFlowTransactionsToCreate,
		); err != nil {
			return errors.Wrap(err, "failed to persist new Lunch Flow transactions")
		}

		// Create actual monetr transactions
		if err := s.repo.InsertTransactions(
			span.Context(),
			transactionsToCreate,
		); err != nil {
			return errors.Wrap(err, "failed to persist new transactions")
		}
	}

	return nil
}

func (s *SyncLunchFlowJob) syncBalances(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := s.log.WithContext(span.Context())

	log.Debug("fetching balance from Lunch Flow")
	balance, err := s.client.GetBalance(
		span.Context(),
		lunch_flow.AccountId(s.bankAccount.LunchFlowBankAccount.LunchFlowId),
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve balance for Lunch Flow sync")
	}

	log.Debug("syncing Lunch Flow balances")

	if balance.Currency != s.bankAccount.Currency {
		log.WithFields(logrus.Fields{
			"currency": logrus.Fields{
				"expected": s.bankAccount.Currency,
				"received": balance.Currency,
			},
		}).Warn("currency returned from Lunch Flow API does not match bank account's currency, this may cause problems!")
	}

	amount, err := currency.ParseFriendlyToAmount(
		balance.Amount.String(),
		s.bankAccount.Currency,
	)
	if err != nil {
		log.WithError(err).Error("failed to parse account balance amount")
		return errors.Wrap(err, "failed to parse Lunch Flow account balance")
	}

	account := *s.bankAccount

	// TODO Log the balance changes and include the new and old balances for
	// debugging reasons.

	// TODO Figure out how available/current should work with Lunch Flow? Since we
	// aren't looking for pending transactions then current is kind of accurate.
	// But if we include pending transactionns in the future then we should
	// subtract their value from the current balance in order to determine the
	// available balance.
	account.CurrentBalance = amount
	account.AvailableBalance = amount

	// TODO UPDATE THE BALANCE ON THE LUNCH FLOW OBJECT TOOOOOO!
	if err := s.repo.UpdateBankAccount(span.Context(), &account); err != nil {
		return errors.Wrap(err, "failed to update bank account balances")
	}

	return nil
}
