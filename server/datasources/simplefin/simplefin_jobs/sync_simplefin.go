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
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
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

type syncSimpleFinContext struct {
	log           *slog.Logger
	args          SyncSimpleFinArguments
	repo          repository.BaseRepository
	secrets       repository.SecretsRepository
	timezone      *time.Location
	client        simplefin.SimpleFinClient
	simpleFinLink *models.SimpleFinLink
	accountSet    *simplefin.AccountSet
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

func (s *syncSimpleFinContext) syncBalances(ctx queue.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	s.log.DebugContext(span.Context(), "syncing SimpleFIN balances")

	bankAccounts, err := s.repo.GetBankAccountsByLinkId(span.Context(), s.args.LinkId)
	if err != nil {
		return err
	}

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
			blog.WarnContext(span.Context(), "SimpleFIN account was not in the response, balances will not be updated")
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

		s.announceProgress(ctx, SimpleFinSyncStatusBalances)
		if err := s.syncBalances(ctx); err != nil {
			return err
		}

		// A successful sync is also what moves a link back from the error status
		// once the user has fixed their token.
		now := ctx.Clock().Now().UTC()
		s.simpleFinLink.Status = models.SimpleFinLinkStatusActive
		s.simpleFinLink.LastSuccessfulUpdate = &now
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
