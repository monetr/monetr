package lunch_flow

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

func SyncLunchFlowCron(ctx queue.Context) error {
	log := ctx.Log()

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
			ctx.Processor(),
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
	queue.Context
	log                   *slog.Logger
	args                  SyncLunchFlowArguments
	repo                  repository.BaseRepository
	secrets               repository.SecretsRepository
	timezone              *time.Location
	client                LunchFlowClient
	bankAccount           *models.BankAccount
	lunchFlowTransactions []Transaction
	existingTransactions  map[string]Transaction
}

func announceProgress(ctx *syncLunchFlowContext, status LunchFlowSyncStatus) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	ctx.log.DebugContext(span.Context(), "sending progress update", "status", status)

	channel := fmt.Sprintf(
		"account:%s:link:%s:bank_account:%s:lunch_flow_sync_progress",
		ctx.args.AccountId, ctx.args.LinkId, ctx.args.BankAccountId,
	)
	j, err := json.Marshal(map[string]any{
		"bankAccountId": ctx.args.BankAccountId,
		"status":        status,
	})

	if err != nil {
		ctx.log.ErrorContext(span.Context(), "failed to encode progress message", "err", err)
		return nil
	}
	if err := ctx.Publisher().Notify(
		ctx,
		ctx.args.AccountId,
		channel,
		string(j),
	); err != nil {
		return errors.Wrap(err, "failed to send progress notification for job")
	}

	return nil
}

func setupClient(ctx *syncLunchFlowContext) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	link, err := ctx.repo.GetLink(span.Context(), ctx.bankAccount.LinkId)
	if err != nil {
		return err
	}

	if link.LunchFlowLink == nil {
		ctx.log.WarnContext(span.Context(), "provided link does not have any Lunch Flow credentials")
		crumbs.IndicateBug(
			span.Context(),
			"BUG: Link was queued to sync with Lunch Flow, but has no Lunch Flow details",
			map[string]any{
				"link": link,
			},
		)
		return nil
	}
	ctx.log = ctx.log.With(
		"linkId", link.LinkId,
		"lunchFlowLinkId", link.LunchFlowLinkId,
	)

	{ // Store the account's timezone information
		account, err := ctx.repo.GetAccount(span.Context())
		if err != nil {
			ctx.log.ErrorContext(span.Context(), "failed to retrieve account for job", "err", err)
			return err
		}

		ctx.timezone, err = account.GetTimezone()
		if err != nil {
			ctx.log.WarnContext(span.Context(), "failed to get account's time zone, defaulting to UTC", "err", err)
			ctx.timezone = time.UTC
		}
	}

	{ // Retrieve the secret and setup the API client
		secret, err := ctx.secrets.Read(span.Context(), link.LunchFlowLink.SecretId)
		if err = errors.Wrap(err, "failed to retrieve credentials for Lunch Flow"); err != nil {
			ctx.log.ErrorContext(span.Context(), "could not retrieve API credentials for Lunch Flow for link", "err", err)
			return err
		}

		client, err := NewLunchFlowClient(
			ctx.log,
			link.LunchFlowLink.ApiUrl,
			secret.Value,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create Lunch Flow API client")
		}

		ctx.client = client
	}

	return nil
}

func SyncLunchFlow(ctx queue.Context, args SyncLunchFlowArguments) error {
	crumbs.IncludeUserInScope(ctx, args.AccountId)
	return ctx.RunInTransaction(ctx, func(innerCtx queue.Context) error {
		span := sentry.SpanFromContext(innerCtx)
		ctx := &syncLunchFlowContext{
			log: innerCtx.Log().With(
				"accountId", args.AccountId,
				"bankAccountId", args.BankAccountId,
			),
			args:                  args,
			Context:               innerCtx,
			lunchFlowTransactions: []Transaction{},
			existingTransactions:  map[string]Transaction{},
		}
		ctx.repo = repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_lunch_flow",
			args.AccountId,
			ctx.DB(),
			ctx.log,
		)
		ctx.secrets = repository.NewSecretsRepository(
			ctx.log,
			ctx.Clock(),
			ctx.DB(),
			ctx.KMS(),
			args.AccountId,
		)

		announceProgress(ctx, LunchFlowSyncStatusBegin)

		{ // Retrieve the bank account from the database, store the details for later
			bankAccount, err := ctx.repo.GetBankAccount(
				ctx,
				args.BankAccountId,
			)
			if err != nil {
				return err
			}

			if bankAccount.LunchFlowBankAccount == nil {
				ctx.log.WarnContext(ctx, "bank account does not have a Lunch Flow account associated with it")
				span.Status = sentry.SpanStatusFailedPrecondition
				return nil
			}
			ctx.log = ctx.log.With(
				slog.Group("lunchFlow",
					"lunchFlowId", bankAccount.LunchFlowBankAccount.LunchFlowId,
					"institutionName", bankAccount.LunchFlowBankAccount.InstitutionName,
				),
			)
			ctx.bankAccount = bankAccount
		}

		if err := setupClient(ctx); err != nil {
			return err
		}

		return nil
	})
}
