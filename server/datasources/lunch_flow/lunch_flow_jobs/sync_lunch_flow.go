package lunch_flow_jobs

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/similar/similar_jobs"
	"github.com/pkg/errors"
)

func LunchFlowSyncNotifcationChannel(
	accountId models.ID[models.Account],
	linkId models.ID[models.Link],
	bankAccountId models.ID[models.BankAccount],
) string {
	return fmt.Sprintf(
		"account:%s:link:%s:bank_account:%s:lunch_flow_sync_progress",
		accountId, linkId, bankAccountId,
	)
}

func SyncLunchFlowCron(ctx queue.Context) error {
	log := ctx.Log()

	if !ctx.Configuration().LunchFlow.Enabled {
		log.InfoContext(ctx, "lunch flow is not enabled")
		return nil
	}

	log.InfoContext(ctx, "retrieving bank accounts to sync with Lunch Flow")

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())

	bankAccounts, err := jobRepo.GetLunchFlowAccountsToSync(ctx)
	if err != nil {
		return err
	}

	if len(bankAccounts) == 0 {
		log.InfoContext(ctx, "no bank accounts to by synced with Lunch Flow at this time")
		return nil
	}

	for _, item := range bankAccounts {
		itemLog := log.With(
			"accountId", item.AccountId,
			"bankAccountId", item.BankAccountId,
			"linkId", item.LinkId,
		)

		itemLog.Log(ctx, logging.LevelTrace, "enqueuing bank account to be synced with Lunch Flow")

		err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			SyncLunchFlow,
			SyncLunchFlowArguments{
				AccountId:     item.AccountId,
				BankAccountId: item.BankAccountId,
				LinkId:        item.LinkId,
			},
		)
		if err != nil {
			itemLog.WarnContext(ctx, "failed to enqueue job to sync with Lunch Flow", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to sync with Lunch Flow", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued bank account to be synced with Lunch Flow")
	}

	return nil
}

type LunchFlowSyncStatus string

const (
	LunchFlowSyncStatusBegin        LunchFlowSyncStatus = "begin"
	LunchFlowSyncStatusTransactions LunchFlowSyncStatus = "transactions"
	LunchFlowSyncStatusBalances     LunchFlowSyncStatus = "balances"
	LunchFlowSyncStatusComplete     LunchFlowSyncStatus = "complete"
	LunchFlowSyncStatusError        LunchFlowSyncStatus = "error"
)

type SyncLunchFlowArguments struct {
	AccountId     models.ID[models.Account]     `json:"accountId"`
	BankAccountId models.ID[models.BankAccount] `json:"bankAccountId"`
	LinkId        models.ID[models.Link]        `json:"linkId"`
}

type syncLunchFlowContext struct {
	log                   *slog.Logger
	args                  SyncLunchFlowArguments
	repo                  repository.BaseRepository
	secrets               repository.SecretsRepository
	timezone              *time.Location
	client                lunch_flow.LunchFlowClient
	bankAccount           *models.BankAccount
	lunchFlowTransactions []lunch_flow.Transaction
	existingTransactions  map[string]models.Transaction
}

func (s *syncLunchFlowContext) announceProgress(
	ctx queue.Context,
	status LunchFlowSyncStatus,
) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "sending progress update", "status", status)

	channel := LunchFlowSyncNotifcationChannel(
		s.args.AccountId,
		s.args.LinkId,
		s.args.BankAccountId,
	)

	j, err := json.Marshal(map[string]any{
		"bankAccountId": s.args.BankAccountId,
		"status":        status,
	})

	if err != nil {
		s.log.ErrorContext(
			span.Context(),
			"failed to encode progress message",
			"err", err,
		)
		return
	}
	if err := ctx.Publisher().Notify(
		ctx,
		s.args.AccountId,
		channel,
		string(j),
	); err != nil {
		s.log.ErrorContext(
			span.Context(),
			"failed to send progress notification",
			"err", err,
		)
	}
}

func (s *syncLunchFlowContext) setupClient(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	link, err := s.repo.GetLink(span.Context(), s.bankAccount.LinkId)
	if err != nil {
		return err
	}

	if link.LunchFlowLink == nil {
		s.log.WarnContext(span.Context(), "provided link does not have any Lunch Flow credentials")
		crumbs.IndicateBug(
			span.Context(),
			"BUG: Link was queued to sync with Lunch Flow, but has no Lunch Flow details",
			map[string]any{
				"link": link,
			},
		)
		return nil
	}
	s.log = s.log.With(
		"linkId", link.LinkId,
		"lunchFlowLinkId", link.LunchFlowLinkId,
	)

	{ // Store the account's timezone information
		account, err := s.repo.GetAccount(span.Context())
		if err != nil {
			s.log.ErrorContext(
				span.Context(),
				"failed to retrieve account for job",
				"err", err,
			)
			return err
		}

		s.timezone, err = account.GetTimezone()
		if err != nil {
			s.log.WarnContext(span.Context(), "failed to get account's time zone, defaulting to UTC", "err", err)
			s.timezone = time.UTC
		}
	}

	{ // Retrieve the secret and setup the API client
		secret, err := s.secrets.Read(span.Context(), link.LunchFlowLink.SecretId)
		if err = errors.Wrap(err, "failed to retrieve credentials for Lunch Flow"); err != nil {
			s.log.ErrorContext(
				span.Context(),
				"could not retrieve API credentials for Lunch Flow for link",
				"err", err,
			)
			return err
		}

		client, err := lunch_flow.NewLunchFlowClient(
			s.log,
			link.LunchFlowLink.ApiUrl,
			secret.Value,
			ctx.Configuration().LunchFlow,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create Lunch Flow API client")
		}

		s.client = client
	}

	return nil
}

func (s *syncLunchFlowContext) hydrateTransactions(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "fetching Lunch Flow transactions")

	var err error
	s.lunchFlowTransactions, err = s.client.GetTransactions(
		span.Context(),
		lunch_flow.LunchFlowAccountId(s.bankAccount.LunchFlowBankAccount.LunchFlowId),
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
		s.log.DebugContext(
			span.Context(),
			"found existing transactions for sync",
			"existingTransactions", count,
		)
	}

	return nil
}

func (s *syncLunchFlowContext) syncTransactions(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "syncing Lunch Flow transactions")

	// transactionsToUpdate := make([]*Transaction, 0)
	lunchFlowTransactionsToCreate := make([]models.LunchFlowTransaction, 0)
	transactionsToCreate := make([]models.Transaction, 0)

	for i := range s.lunchFlowTransactions {
		externalTransaction := s.lunchFlowTransactions[i]
		tlog := s.log.With(
			slog.Group("lunchFlowTransaction",
				"lunchFlowId", externalTransaction.Id,
				"amount", externalTransaction.Amount.String(),
				"currency", externalTransaction.Currency,
			),
		)

		if externalTransaction.Currency != s.bankAccount.Currency {
			tlog.WarnContext(span.Context(),
				"currency on transaction does not match bank account's stored currency, this may cause problems!",
				slog.Group("currency",
					"expected", s.bankAccount.Currency,
					"received", externalTransaction.Currency,
				),
			)
		}

		amount, err := currency.ParseFriendlyToAmount(
			externalTransaction.Amount.String(),
			s.bankAccount.Currency,
		)
		if err != nil {
			tlog.ErrorContext(span.Context(), "failed to parse transaction amount", "err", err)
			continue
		}
		// Invert the amount for monetr!
		amount *= -1

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

		description := myownsanity.CoalesceStrings(
			externalTransaction.Description,
			externalTransaction.Merchant,
			fallbackName,
		)

		date, err := lunch_flow.ParseDate(externalTransaction.Date, s.timezone)
		if err != nil {
			s.log.ErrorContext(span.Context(),
				"failed to parse transaction date from Lunch Flow",
				"err", err,
				"value", externalTransaction.Date,
			)
			continue
		}

		transaction, ok := s.existingTransactions[externalTransaction.Id]
		if !ok {
			lunchFlowTransaction := models.LunchFlowTransaction{
				LunchFlowTransactionId: models.NewID[models.LunchFlowTransaction](),
				AccountId:              s.bankAccount.AccountId,
				LunchFlowBankAccountId: *s.bankAccount.LunchFlowBankAccountId,
				LunchFlowId:            externalTransaction.Id,
				Merchant:               externalTransaction.Merchant,
				Description:            description,
				Date:                   date,
				Currency:               externalTransaction.Currency,
				Amount:                 amount,
				IsPending:              false,
			}

			transaction = models.Transaction{
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
		tlog.Log(
			span.Context(),
			logging.LevelTrace,
			"transaction from Lunch Flow already exists in monetr, skipping",
		)
		_ = transaction
		continue
	}

	// Persist any new transactions.
	if count := len(transactionsToCreate); count > 0 {
		s.log.InfoContext(
			span.Context(),
			"creating new Lunch Flow transactions",
			"new", count,
		)

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

func (s *syncLunchFlowContext) syncBalances(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "fetching balance from Lunch Flow")
	balance, err := s.client.GetBalance(
		span.Context(),
		lunch_flow.LunchFlowAccountId(s.bankAccount.LunchFlowBankAccount.LunchFlowId),
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve balance for Lunch Flow sync")
	}

	s.log.DebugContext(span.Context(), "syncing Lunch Flow balances")

	if balance.Currency != s.bankAccount.Currency {
		s.log.WarnContext(span.Context(),
			"currency returned from Lunch Flow API does not match bank account's currency, this may cause problems!",
			slog.Group("currency",
				"expected", s.bankAccount.Currency,
				"received", balance.Currency,
			),
		)
	}

	amount, err := currency.ParseFriendlyToAmount(
		balance.Amount.String(),
		s.bankAccount.Currency,
	)
	if err != nil {
		s.log.ErrorContext(
			span.Context(),
			"failed to parse account balance amount",
			"err", err,
		)
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

	if err := s.repo.UpdateBankAccount(span.Context(), &account); err != nil {
		return errors.Wrap(err, "failed to update bank account balances")
	}

	lunchFlowBankAccount, err := s.repo.GetLunchFlowBankAccount(
		span.Context(),
		*account.LunchFlowBankAccountId,
	)
	if err != nil {
		return errors.Wrap(err, "failed to read Lunch Flow bank account for balance update")
	}

	lunchFlowBankAccount.CurrentBalance = amount

	if err := s.repo.UpdateLunchFlowBankAccount(
		span.Context(),
		lunchFlowBankAccount,
	); err != nil {
		return errors.Wrap(err, "failed to update Lunch Flow bank account balance")
	}

	return nil
}

func SyncLunchFlow(ctx queue.Context, args SyncLunchFlowArguments) error {
	if !ctx.Configuration().LunchFlow.Enabled {
		ctx.Log().InfoContext(ctx, "lunch flow is not enabled")
		return nil
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		span := sentry.SpanFromContext(ctx)
		s := &syncLunchFlowContext{
			log: ctx.Log().With(
				"accountId", args.AccountId,
				"bankAccountId", args.BankAccountId,
			),
			args:                  args,
			lunchFlowTransactions: []lunch_flow.Transaction{},
			existingTransactions:  map[string]models.Transaction{},
		}
		s.repo = repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_lunch_flow",
			args.AccountId,
			ctx.DB(),
			s.log,
		)
		s.secrets = repository.NewSecretsRepository(
			s.log,
			ctx.Clock(),
			ctx.DB(),
			ctx.KMS(),
			args.AccountId,
		)

		s.announceProgress(ctx, LunchFlowSyncStatusBegin)

		{ // Retrieve the bank account from the database, store the details for later
			bankAccount, err := s.repo.GetBankAccount(
				ctx,
				args.BankAccountId,
			)
			if err != nil {
				return err
			}

			if bankAccount.LunchFlowBankAccount == nil {
				s.log.WarnContext(
					ctx,
					"bank account does not have a Lunch Flow account associated with it",
				)
				span.Status = sentry.SpanStatusFailedPrecondition
				return nil
			}
			s.log = s.log.With(
				slog.Group("lunchFlow",
					"lunchFlowId", bankAccount.LunchFlowBankAccount.LunchFlowId,
					"institutionName", bankAccount.LunchFlowBankAccount.InstitutionName,
				),
			)
			s.bankAccount = bankAccount
		}

		if err := s.setupClient(ctx); err != nil {
			return err
		}

		s.announceProgress(ctx, LunchFlowSyncStatusTransactions)
		if err := s.hydrateTransactions(ctx); err != nil {
			return err
		}

		if err := s.syncTransactions(ctx); err != nil {
			return err
		}

		s.announceProgress(ctx, LunchFlowSyncStatusBalances)
		if err := s.syncBalances(ctx); err != nil {
			return err
		}

		// Also kick off the transaction similarity job.
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			similar_jobs.CalculateTransactionClusters,
			similar_jobs.CalculateTransactionClustersArguments{
				AccountId:     s.args.AccountId,
				BankAccountId: s.args.BankAccountId,
			},
		); err != nil {
			s.log.ErrorContext(
				ctx,
				"failed to queue transaction cluster calculations",
				"err", err,
			)
		}

		s.announceProgress(ctx, LunchFlowSyncStatusComplete)
		s.log.InfoContext(ctx, "finished syncing Lunch Flow account")

		return nil
	})
}
