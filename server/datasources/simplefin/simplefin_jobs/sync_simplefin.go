package simplefin_jobs

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/datasources/simplefin"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/similar/similar_jobs"
	"github.com/monetr/monetr/server/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

func SimpleFinSyncNotificationChannel(
	accountId models.ID[models.Account],
	linkId models.ID[models.Link],
) string {
	return fmt.Sprintf(
		"account:%s:link:%s:simplefin_sync_progress",
		accountId, linkId,
	)
}

type SimpleFinSyncStatus string

const (
	SimpleFinSyncStatusBegin        SimpleFinSyncStatus = "begin"
	SimpleFinSyncStatusTransactions SimpleFinSyncStatus = "transactions"
	SimpleFinSyncStatusBalances     SimpleFinSyncStatus = "balances"
	SimpleFinSyncStatusComplete     SimpleFinSyncStatus = "complete"
	SimpleFinSyncStatusError        SimpleFinSyncStatus = "error"
)

type SyncSimpleFinArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
	LinkId    models.ID[models.Link]    `json:"linkId"`
}

func SyncSimpleFinCron(ctx queue.Context) error {
	log := ctx.Log()

	if !ctx.Configuration().SimpleFin.IsEnabled() {
		log.InfoContext(ctx, "simplefin is not enabled")
		return nil
	}

	log.InfoContext(ctx, "retrieving links to sync with SimpleFIN")

	jobRepo := repository.NewJobRepository(ctx.DB(), ctx.Clock())

	links, err := jobRepo.GetSimpleFinLinksToSync(ctx)
	if err != nil {
		return err
	}

	if len(links) == 0 {
		log.InfoContext(ctx, "no links to be synced with SimpleFIN at this time")
		return nil
	}

	if err = queue.BulkEnqueue(
		ctx,
		ctx.Enqueuer(),
		SyncSimpleFin,
		myownsanity.Map(
			links,
			func(item models.Link) SyncSimpleFinArguments {
				return SyncSimpleFinArguments{
					AccountId: item.AccountId,
					LinkId:    item.LinkId,
				}
			}),
	); err != nil {
		log.WarnContext(ctx, "failed to enqueue jobs to sync with SimpleFIN", "err", err)
		crumbs.Warn(ctx, "Failed to enqueue jobs to sync with SimpleFIN", "job", map[string]any{
			"error": err,
		})
		return err
	}

	return nil
}

type syncSimpleFinContext struct {
	log           *slog.Logger
	args          SyncSimpleFinArguments
	repo          repository.BaseRepository
	secrets       repository.SecretsRepository
	timezone      *time.Location
	client        simplefin.SimpleFinClient
	simpleFinLink *models.SimpleFinLink
	accountSet    *simplefin.AccountSet
	bankAccounts  []simpleFinBankAccountMatch
	// hasChanges is set when the sync creates a transaction or changes a balance,
	// the last successful update is only bumped when something actually changed.
	hasChanges bool
}

type simpleFinBankAccountMatch struct {
	log         *slog.Logger
	bankAccount models.BankAccount
	external    simplefin.Account
}

func (s *syncSimpleFinContext) announceProgress(
	ctx queue.Context,
	status SimpleFinSyncStatus,
) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "sending progress update", "status", status)

	channel := SimpleFinSyncNotificationChannel(
		s.args.AccountId,
		s.args.LinkId,
	)

	j, err := json.Marshal(map[string]any{
		"linkId": s.args.LinkId,
		"status": status,
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

func (s *syncSimpleFinContext) setupClient(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

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
		secret, err := s.secrets.Read(span.Context(), s.simpleFinLink.SecretId)
		if err != nil {
			s.log.ErrorContext(
				span.Context(),
				"failed to retrieve credentials for SimpleFIN",
				"err", err,
			)
			return errors.Wrap(err, "failed to retrieve credentials for SimpleFIN")
		}

		client, err := simplefin.NewSimpleFinClient(
			s.log,
			secret.Value,
			ctx.Configuration().SimpleFin,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create SimpleFIN API client")
		}

		s.client = client
	}

	return nil
}

// matchBankAccounts pairs each of the link's bank accounts with its account
// in the SimpleFIN response. Only bank accounts the user has turned on and that
// are actually in the response are kept, both the transaction and balance steps
// work off of this.
func (s *syncSimpleFinContext) matchBankAccounts(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	bankAccounts, err := s.repo.GetBankAccountsByLinkId(span.Context(), s.args.LinkId)
	if err != nil {
		return err
	}

	s.bankAccounts = make([]simpleFinBankAccountMatch, 0, len(bankAccounts))
	// The SimpleFIN account ID is only unique within its connection, so we need
	// to match on both
	for _, joined := range myownsanity.LeftJoin(
		bankAccounts,
		s.accountSet.Accounts,
		func(internal models.BankAccount, external simplefin.Account) bool {
			return internal.SimpleFinBankAccount != nil &&
				internal.SimpleFinBankAccount.SimpleFinId == external.Id &&
				internal.SimpleFinBankAccount.ConnectionId == external.ConnectionId
		},
	) {
		bankAccount := joined.From

		// Only sync the bank accounts the user has turned on
		if bankAccount.SimpleFinBankAccount == nil {
			continue
		}
		if bankAccount.SimpleFinBankAccount.Status != models.SimpleFinBankAccountStatusActive {
			continue
		}
		if bankAccount.DeletedAt != nil {
			continue
		}
		if bankAccount.SimpleFinBankAccount.DeletedAt != nil {
			continue
		}

		blog := s.log.With(
			"bankAccountId", bankAccount.BankAccountId,
			slog.Group("simplefin",
				"simpleFinId", bankAccount.SimpleFinBankAccount.SimpleFinId,
				"connectionId", bankAccount.SimpleFinBankAccount.ConnectionId,
			),
		)

		// This is usually because SimpleFIN couldn't reach the institution, there
		// should be an errlist entry for the connection that explains why.
		if len(joined.Join) == 0 {
			blog.WarnContext(span.Context(), "SimpleFIN account was not in the response, it will not be synced")
			continue
		}
		external := joined.Join[0]

		if external.Currency != bankAccount.Currency {
			blog.WarnContext(span.Context(),
				"currency returned from SimpleFIN does not match bank account's currency, this may cause problems!",
				slog.Group("currency",
					"expected", bankAccount.Currency,
					"received", external.Currency,
				),
			)
		}

		s.bankAccounts = append(s.bankAccounts, simpleFinBankAccountMatch{
			log:         blog,
			bankAccount: bankAccount,
			external:    external,
		})
	}

	return nil
}

func (s *syncSimpleFinContext) syncTransactions(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "syncing SimpleFIN transactions")

	for _, match := range s.bankAccounts {
		bankAccount := match.bankAccount
		external := match.external
		blog := match.log

		transactionCurrency, err := currency.GetCurrency("en", bankAccount.Currency)
		if err != nil {
			blog.WarnContext(span.Context(), "bank account has an unsupported currency, transactions will not be synced",
				"currency", bankAccount.Currency,
				"err", err,
			)
			continue
		}

		simpleFinIds := make([]string, len(external.Transactions))
		for i := range external.Transactions {
			simpleFinIds[i] = external.Transactions[i].Id
		}

		existingTransactions, err := s.repo.GetTransactionsBySimpleFinId(
			span.Context(),
			bankAccount.BankAccountId,
			simpleFinIds,
		)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve existing SimpleFIN transactions")
		}

		if count := len(existingTransactions); count > 0 {
			blog.DebugContext(
				span.Context(),
				"found existing transactions for sync",
				"existingTransactions", count,
			)
		}

		simpleFinTransactionsToCreate := make([]models.SimpleFinTransaction, 0)
		transactionsToCreate := make([]models.Transaction, 0)
		// If SimpleFIN ever sends the same transaction twice in one response then
		// inserting both would fail the unique constraint, and every sync after
		// this one would fail the same way.
		seen := make(map[string]struct{}, len(external.Transactions))
		for _, externalTransaction := range external.Transactions {
			tlog := blog.With(
				slog.Group("simpleFinTransaction",
					"simpleFinId", externalTransaction.Id,
					"amount", externalTransaction.Amount.String(),
				),
			)

			// We never ask for pending transactions so these shouldn't show up, but
			// if they do we don't want them.
			if externalTransaction.Pending {
				tlog.DebugContext(span.Context(), "skipping pending SimpleFIN transaction")
				continue
			}

			if _, ok := seen[externalTransaction.Id]; ok {
				tlog.WarnContext(span.Context(), "SimpleFIN returned the same transaction more than once, skipping")
				continue
			}
			seen[externalTransaction.Id] = struct{}{}

			if _, ok := existingTransactions[externalTransaction.Id]; ok {
				// TODO Handle updating transactions too!
				tlog.Log(
					span.Context(),
					logging.LevelTrace,
					"transaction from SimpleFIN already exists in monetr, skipping",
				)
				continue
			}

			// Amounts are JSON numbers so they always use . for the decimal and never
			// have grouping, which is what en gives us.
			amount, err := currency.ParseCurrency(
				externalTransaction.Amount.String(),
				transactionCurrency,
			)
			if err != nil {
				tlog.ErrorContext(span.Context(), "failed to parse transaction amount", "err", err)
				continue
			}
			// SimpleFIN is positive for deposits, monetr is positive for debits.
			amount *= -1

			// Transacted at is optional and what it means varies by institution, so
			// posted is what we use for the date.
			posted := time.Unix(externalTransaction.Posted, 0).UTC()
			date := util.Midnight(posted, s.timezone)
			var transactedAt *time.Time
			if externalTransaction.TransactedAt != nil {
				transactedAt = new(time.Unix(*externalTransaction.TransactedAt, 0).UTC())
			}

			name := myownsanity.CoalesceStrings(
				externalTransaction.Description,
				fmt.Sprintf(
					"Transaction %s %s %s",
					date.Format(time.DateOnly),
					externalTransaction.Amount.String(),
					external.Currency,
				),
			)

			simpleFinTransaction := models.SimpleFinTransaction{
				SimpleFinTransactionId: models.NewID[models.SimpleFinTransaction](),
				AccountId:              bankAccount.AccountId,
				SimpleFinBankAccountId: *bankAccount.SimpleFinBankAccountId,
				SimpleFinId:            externalTransaction.Id,
				Description:            name,
				Amount:                 amount,
				Posted:                 posted,
				TransactedAt:           transactedAt,
			}

			transaction := models.Transaction{
				AccountId:              bankAccount.AccountId,
				BankAccountId:          bankAccount.BankAccountId,
				SimpleFinTransactionId: &simpleFinTransaction.SimpleFinTransactionId,
				SimpleFinTransaction:   &simpleFinTransaction,
				Amount:                 amount,
				Date:                   date,
				Name:                   name,
				OriginalName:           name,
				IsPending:              false,
				Source:                 models.TransactionSourceSimpleFin,
			}

			simpleFinTransactionsToCreate = append(simpleFinTransactionsToCreate, simpleFinTransaction)
			transactionsToCreate = append(transactionsToCreate, transaction)
		}

		if count := len(transactionsToCreate); count > 0 {
			blog.InfoContext(
				span.Context(),
				"creating new SimpleFIN transactions",
				"new", count,
			)

			// Create SimpleFIN transactions first
			if err := s.repo.CreateSimpleFinTransactions(
				span.Context(),
				simpleFinTransactionsToCreate,
			); err != nil {
				return errors.Wrap(err, "failed to persist new SimpleFIN transactions")
			}

			// Create actual monetr transactions
			if err := s.repo.InsertTransactions(
				span.Context(),
				transactionsToCreate,
			); err != nil {
				return errors.Wrap(err, "failed to persist new transactions")
			}

			s.hasChanges = true
		}
	}

	return nil
}

func (s *syncSimpleFinContext) syncBalances(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "syncing SimpleFIN balances")

	for _, match := range s.bankAccounts {
		bankAccount := match.bankAccount
		external := match.external
		blog := match.log

		balanceCurrency, err := currency.GetCurrency("en", bankAccount.Currency)
		if err != nil {
			blog.WarnContext(span.Context(), "bank account has an unsupported currency, balances will not be updated",
				"currency", bankAccount.Currency,
				"err", err,
			)
			continue
		}

		// The balances are JSON numbers so they always use . for the decimal and
		// never have grouping, which is what en gives us. If one account has a bad
		// balance we skip it instead of failing the whole sync.
		currentBalance, err := currency.ParseCurrency(
			external.Balance.String(),
			balanceCurrency,
		)
		if err != nil {
			blog.WarnContext(span.Context(), "failed to parse SimpleFIN account balance, balances will not be updated", "err", err)
			continue
		}

		// Available balance is optional, if it isn't there then just use the
		// current balance
		availableBalance := currentBalance
		if external.AvailableBalance != nil {
			availableBalance, err = currency.ParseCurrency(
				external.AvailableBalance.String(),
				balanceCurrency,
			)
			if err != nil {
				blog.WarnContext(span.Context(), "failed to parse SimpleFIN account available balance, balances will not be updated", "err", err)
				continue
			}
		}

		if bankAccount.CurrentBalance != currentBalance ||
			bankAccount.AvailableBalance != availableBalance {
			s.hasChanges = true
		}

		bankAccount.CurrentBalance = currentBalance
		bankAccount.AvailableBalance = availableBalance
		if err := s.repo.UpdateBankAccount(span.Context(), &bankAccount); err != nil {
			return errors.Wrap(err, "failed to update bank account balances")
		}

		simpleFinBankAccount := bankAccount.SimpleFinBankAccount
		simpleFinBankAccount.CurrentBalance = currentBalance
		simpleFinBankAccount.AvailableBalance = availableBalance
		if err := s.repo.UpdateSimpleFinBankAccount(
			span.Context(),
			simpleFinBankAccount,
		); err != nil {
			return err
		}
	}

	return nil
}

func SyncSimpleFin(ctx queue.Context, args SyncSimpleFinArguments) error {
	if !ctx.Configuration().SimpleFin.IsEnabled() {
		ctx.Log().InfoContext(ctx, "simplefin is not enabled")
		return nil
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	// Record the attempt outside of the sync transaction so that it persists even
	// if the sync fails, otherwise the cron would retry this link on every tick.
	if err := recordSimpleFinSyncAttempt(ctx, args); err != nil {
		return err
	}

	s := &syncSimpleFinContext{
		log: ctx.Log().With(
			"accountId", args.AccountId,
			"linkId", args.LinkId,
		),
		args: args,
	}
	s.repo = repository.NewRepositoryFromSession(
		ctx.Clock(),
		"user_simplefin",
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

	s.announceProgress(ctx, SimpleFinSyncStatusBegin)

	{ // Make sure the link is still something we should be syncing
		link, err := s.repo.GetLink(ctx, args.LinkId)
		if err != nil {
			return err
		}

		if link.SimpleFinLink == nil {
			s.log.WarnContext(ctx, "provided link does not have any SimpleFIN credentials")
			crumbs.IndicateBug(
				ctx,
				"BUG: Link was queued to sync with SimpleFIN, but has no SimpleFIN details",
				map[string]any{
					"link": link,
				},
			)
			return nil
		}

		s.simpleFinLink = link.SimpleFinLink
		s.log = s.log.With("simpleFinLinkId", link.SimpleFinLink.SimpleFinLinkId)

		switch s.simpleFinLink.Status {
		case models.SimpleFinLinkStatusPending, models.SimpleFinLinkStatusDeactivated:
			// The cron only picks active links and the manual sync allows active or
			// error, so this only happens if the status changed after the job was
			// enqueued.
			s.log.WarnContext(ctx, "SimpleFIN link is not in a status that can be synced, skipping",
				"status", s.simpleFinLink.Status,
			)
			return nil
		}
	}

	if err := s.setupClient(ctx); err != nil {
		return err
	}

	// Only count the syncs that actually make a request to SimpleFIN, that way
	// this lines up with how much of the user's quota we're using. Anything that
	// doesn't set the result is a failure, including FailWithoutRetry since that
	// panics.
	result := "failure"
	defer func() {
		simpleFinSyncs.With(prometheus.Labels{
			"result": result,
		}).Inc()
	}()

	// The request to SimpleFIN is outside of the transaction because it can take
	// up to 60 seconds and we don't want to hold a transaction open that long.
	// This way the error status we set below for a 402 or 403 also doesn't get
	// rolled back.
	accountSet, err := s.client.GetAccounts(ctx, simplefin.GetAccountsOptions{
		// Every request costs the same against the quota no matter how big the
		// window is, so always ask for the full 90 days and let the dedupe sort it
		// out. This also back fills accounts that were turned on later.
		StartDate: ctx.Clock().Now().AddDate(0, 0, -90),
	})
	var deniedError *models.SimpleFinLinkError
	switch errors.Cause(err) {
	case nil:
		// Do nothing
	case simplefin.ErrPaymentRequired:
		deniedError = &models.SimpleFinLinkError{
			Code:    "payment_required",
			Message: "SimpleFIN requires payment, check your SimpleFIN account",
		}
	case simplefin.ErrForbidden:
		deniedError = &models.SimpleFinLinkError{
			Code:    "forbidden",
			Message: "SimpleFIN access was denied, the token may have been revoked",
		}
	default:
		// Retrying would spend another request out of the user's quota, the next
		// cron run will try again on its own.
		s.log.ErrorContext(ctx, "failed to retrieve accounts from SimpleFIN", "err", err)
		s.announceProgress(ctx, SimpleFinSyncStatusError)
		return queue.FailWithoutRetry(ctx, errors.Wrap(err, "failed to retrieve accounts from SimpleFIN"))
	}

	// If SimpleFIN wants to be paid or the token was revoked then the token is
	// dead until the user fixes it. Put the link in the error status so the cron
	// stops picking it up and the user can see why.
	if deniedError != nil {
		s.log.WarnContext(ctx, "SimpleFIN denied access for link", "code", deniedError.Code)
		s.simpleFinLink.Status = models.SimpleFinLinkStatusError
		s.simpleFinLink.LastErrors = []models.SimpleFinLinkError{
			*deniedError,
		}
		if err := s.repo.UpdateSimpleFinLink(ctx, s.simpleFinLink); err != nil {
			s.log.ErrorContext(ctx, "failed to update SimpleFIN link after access was denied", "err", err)
			s.announceProgress(ctx, SimpleFinSyncStatusError)
			return queue.FailWithoutRetry(ctx, err)
		}

		s.announceProgress(ctx, SimpleFinSyncStatusError)
		result = "denied"
		return nil
	}

	s.accountSet = accountSet

	if err := ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		s.repo = repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_simplefin",
			args.AccountId,
			ctx.DB(),
			s.log,
		)

		// The deprecated errors field is where the Bridge puts things like quota
		// warnings, so we want those to show up in the logs.
		if len(s.accountSet.Errors) > 0 {
			s.log.WarnContext(ctx, "SimpleFIN returned errors with accounts",
				"errors", s.accountSet.Errors,
			)
		}

		// Always replace the errors on the link, even when there aren't any, so old
		// errors don't stick around after they've been resolved.
		s.simpleFinLink.LastErrors = s.accountSet.GetLinkErrors()

		if err := s.matchBankAccounts(ctx); err != nil {
			return err
		}

		s.announceProgress(ctx, SimpleFinSyncStatusTransactions)
		if err := s.syncTransactions(ctx); err != nil {
			return err
		}

		s.announceProgress(ctx, SimpleFinSyncStatusBalances)
		if err := s.syncBalances(ctx); err != nil {
			return err
		}

		// A successful sync is also what moves a link back from the error status
		// once the user has fixed their token.
		now := ctx.Clock().Now().UTC()
		s.simpleFinLink.Status = models.SimpleFinLinkStatusActive
		// Only bump the last successful update when we actually got new data, that
		// way it reflects when the data last changed rather than when we last
		// asked. If it's never been set then set it now so the link doesn't look
		// like it has never synced.
		if s.hasChanges || s.simpleFinLink.LastSuccessfulUpdate == nil {
			s.simpleFinLink.LastSuccessfulUpdate = &now
		}
		s.simpleFinLink.LastAttemptedUpdate = &now
		if err := s.repo.UpdateSimpleFinLink(ctx, s.simpleFinLink); err != nil {
			s.log.ErrorContext(ctx, "failed to update SimpleFIN link after sync", "err", err)
			return err
		}

		return nil
	}); err != nil {
		// Same as above, a retry would fetch from SimpleFIN again.
		s.log.ErrorContext(ctx, "failed to sync SimpleFIN link", "err", err)
		s.announceProgress(ctx, SimpleFinSyncStatusError)
		return queue.FailWithoutRetry(ctx, err)
	}

	// Kick off the transaction similarity job for each bank account we synced.
	// This is after the commit so the job can see the new transactions, and a
	// failure here shouldn't fail a sync that already succeeded.
	if len(s.bankAccounts) > 0 {
		if err := queue.BulkEnqueue(
			ctx,
			ctx.Enqueuer(),
			similar_jobs.CalculateTransactionClusters,
			myownsanity.Map(
				s.bankAccounts,
				func(match simpleFinBankAccountMatch) similar_jobs.CalculateTransactionClustersArguments {
					return similar_jobs.CalculateTransactionClustersArguments{
						AccountId:     s.args.AccountId,
						BankAccountId: match.bankAccount.BankAccountId,
					}
				},
			),
		); err != nil {
			s.log.ErrorContext(
				ctx,
				"failed to queue transaction cluster calculations",
				"err", err,
			)
		}
	}

	s.announceProgress(ctx, SimpleFinSyncStatusComplete)
	s.log.InfoContext(ctx, "finished syncing SimpleFIN link")
	result = "success"

	return nil
}

func recordSimpleFinSyncAttempt(ctx queue.Context, args SyncSimpleFinArguments) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	repo := repository.NewRepositoryFromSession(
		ctx.Clock(),
		"user_simplefin",
		args.AccountId,
		ctx.DB(),
		ctx.Log(),
	)

	link, err := repo.GetLink(span.Context(), args.LinkId)
	if err != nil {
		return err
	}

	// If the link isn't a SimpleFIN link then there is nothing to record, the
	// sync itself will handle that case.
	if link.SimpleFinLink == nil {
		return nil
	}

	link.SimpleFinLink.LastAttemptedUpdate = new(ctx.Clock().Now().UTC())
	if err := repo.UpdateSimpleFinLink(span.Context(), link.SimpleFinLink); err != nil {
		return errors.Wrap(err, "failed to record SimpleFIN sync attempt")
	}

	return nil
}
