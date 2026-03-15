package plaid_jobs

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type SyncPlaidAccountsArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
	LinkId    models.ID[models.Link]    `json:"linkId"`
}

type syncPlaidAccountsJob struct {
	log     *slog.Logger
	repo    repository.BaseRepository
	secrets repository.SecretsRepository

	bankAccounts      []models.BankAccount
	plaidBankAccounts []platypus.BankAccount
}

// findMissingAccounts will return a map of all the bank accounts who are
// currently in an active status, but who are no longer being returned by Plaid.
// These accounts will be marked as inactive.
func (j *syncPlaidAccountsJob) findMissingAccounts(
	ctx context.Context,
) (missingAccounts map[models.ID[models.BankAccount]]models.BankAccount) {
	missingAccounts = make(map[models.ID[models.BankAccount]]models.BankAccount)
MissingAccounts:
	for x := range j.bankAccounts {
		bankAccount := j.bankAccounts[x]
		for y := range j.plaidBankAccounts {
			plaidBankAccount := j.plaidBankAccounts[y]
			if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
				j.log.DebugContext(
					ctx,
					"bank account is still present in plaid and is considered active",
					"bankAccountId", bankAccount.BankAccountId,
					"plaid_bankAccountId", bankAccount.PlaidBankAccount.PlaidId,
				)
				// TODO Check bank account status here too, if the status is inactive
				// but we see the account again then that means it is active again.
				continue MissingAccounts
			}
		}

		if bankAccount.Status == models.BankAccountStatusInactive {
			// Bank account is already considered missing, skip it.
			j.log.Log(
				ctx,
				logging.LevelTrace,
				"bank account is already inactive, it does not need to be updated",
				"bankAccountId", bankAccount.BankAccountId,
				"plaid_bankAccountId", bankAccount.PlaidBankAccount.PlaidId,
			)
			continue
		}

		j.log.InfoContext(
			ctx,
			"bank account is no longer present in plaid and is considered inactive",
			"bankAccountId", bankAccount.BankAccountId,
			"plaid_bankAccountId", bankAccount.PlaidBankAccount.PlaidId,
		)
		missingAccounts[bankAccount.BankAccountId] = bankAccount
	}

	return missingAccounts
}

// findActiveAccounts will return a map of all the accounts who were previously
// marked as inactive but are now being seen in plaid's API responses again.
// This would be extremely unusual.
func (j *syncPlaidAccountsJob) findActiveAccounts(
	ctx context.Context,
) (activeAcounts map[models.ID[models.BankAccount]]models.BankAccount) {
	activeAcounts = make(map[models.ID[models.BankAccount]]models.BankAccount)
ActiveAccounts:
	for x := range j.bankAccounts {
		bankAccount := j.bankAccounts[x]

		// If the account is already marked as active then skip it.
		if bankAccount.Status == models.BankAccountStatusActive {
			continue ActiveAccounts
		}

		for y := range j.plaidBankAccounts {
			plaidBankAccount := j.plaidBankAccounts[y]
			if plaidBankAccount.GetAccountId() == bankAccount.PlaidBankAccount.PlaidId {
				activeAcounts[bankAccount.BankAccountId] = bankAccount
				j.log.InfoContext(
					ctx,
					"found inactive account that is present in Plaid again, will be updated to show as active",
					"bankAccountId", bankAccount.BankAccountId,
					"plaid_bankAccountId", bankAccount.PlaidBankAccount.PlaidId,
				)
				continue ActiveAccounts
			}
		}
	}

	return activeAcounts
}

func SyncPlaidAccountsCron(ctx queue.Context) error {
	log := ctx.Log()

	log.InfoContext(ctx, "retrieving links to sync with Plaid for updated accounts")

	links := make([]models.Link, 0)
	cutoff := ctx.Clock().Now().AddDate(0, 0, -7)
	err := ctx.DB().ModelContext(ctx, &links).
		Join(`INNER JOIN "plaid_links" AS "plaid_link"`).
		JoinOn(`"plaid_link"."plaid_link_id" = "link"."plaid_link_id"`).
		Where(`"plaid_link"."status" = ?`, models.PlaidLinkStatusSetup).
		Where(`"plaid_link"."last_account_sync" < ? OR "plaid_link"."last_account_sync" IS NULL`, cutoff).
		Where(`"plaid_link"."deleted_at" IS NULL`).
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Where(`"link"."deleted_at" IS NULL`).
		Select(&links)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve links that need to be synced with plaid for updated accounts")
	}

	if len(links) == 0 {
		log.DebugContext(ctx, "no plaid links need to be synced at this time for updating accounts")
		return nil
	}

	log.InfoContext(ctx, "syncing plaid links for accounts", "count", len(links))

	for _, item := range links {
		itemLog := log.With(
			"accountId", item.AccountId,
			"linkId", item.LinkId,
		)
		itemLog.Log(ctx, logging.LevelTrace, "enqueuing link to be synced with plaid for accounts")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			SyncPlaidAccounts,
			SyncPlaidAccountsArguments{
				AccountId: item.AccountId,
				LinkId:    item.LinkId,
			},
		); err != nil {
			itemLog.WarnContext(ctx, "failed to enqueue job to sync with plaid accounts", "err", err)
			crumbs.Warn(ctx, "Failed to enqueue job to sync with plaid accounts", "job", map[string]any{
				"error": err,
			})
			continue
		}

		itemLog.Log(ctx, logging.LevelTrace, "successfully enqueued link to be synced with plaid accounts")
	}

	return nil
}

func SyncPlaidAccounts(ctx queue.Context, args SyncPlaidAccountsArguments) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		span := sentry.SpanFromContext(ctx)
		crumbs.AddTag(ctx, "accountId", args.AccountId.String())
		crumbs.AddTag(ctx, "linkId", args.LinkId.String())

		log := ctx.Log().With(
			"accountId", args.AccountId,
			"linkId", args.LinkId,
		)

		j := &syncPlaidAccountsJob{
			log: log,
			repo: repository.NewRepositoryFromSession(
				ctx.Clock(),
				"user_plaid",
				args.AccountId,
				ctx.DB(),
				log,
			),
			secrets: repository.NewSecretsRepository(
				log,
				ctx.Clock(),
				ctx.DB(),
				ctx.KMS(),
				args.AccountId,
			),
			bankAccounts:      []models.BankAccount{},
			plaidBankAccounts: []platypus.BankAccount{},
		}

		link, err := j.repo.GetLink(ctx, args.LinkId)
		if err = errors.Wrap(err, "failed to retrieve link to sync with plaid"); err != nil {
			log.ErrorContext(ctx, "cannot sync without link", "err", err)
			return err
		}

		if link.PlaidLink == nil {
			log.WarnContext(ctx, "provided link does not have any plaid credentials")
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

		log = log.With(
			"plaidLinkId", link.PlaidLink.PlaidLinkId,
			"plaid_institutionId", link.PlaidLink.InstitutionId,
			"plaid_institutionName", link.PlaidLink.InstitutionName,
			"plaid_itemId", link.PlaidLink.PlaidId,
		)

		// This way other methods will have these log fields too.
		j.log = log

		plaidLink := link.PlaidLink

		j.bankAccounts, err = j.repo.GetBankAccountsWithPlaidByLinkId(
			ctx,
			link.LinkId,
		)
		if err = errors.Wrap(err, "failed to read bank accounts for plaid sync"); err != nil {
			log.ErrorContext(ctx, "cannot sync without bank accounts", "err", err)
			return err
		}
		crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.PlaidId)
		crumbs.AddTag(ctx, "plaid.institution_id", link.PlaidLink.InstitutionId)
		crumbs.AddTag(ctx, "plaid.institution_name", link.PlaidLink.InstitutionName)

		if len(j.bankAccounts) == 0 {
			log.WarnContext(ctx, "no bank accounts for plaid link")
			crumbs.Debug(ctx, "No bank accounts setup for plaid link", nil)
			return nil
		}

		secret, err := j.secrets.Read(ctx, plaidLink.SecretId)
		if err = errors.Wrap(err, "failed to retrieve access token for plaid link"); err != nil {
			log.ErrorContext(ctx, "could not retrieve API credentials for Plaid for link, this job will be retried", "err", err)
			return err
		}

		plaidClient, err := ctx.Platypus().NewClient(
			ctx,
			link,
			secret.Value,
			plaidLink.PlaidId,
		)
		if err != nil {
			log.ErrorContext(ctx, "failed to create plaid client for link", "err", err)
			return err
		}

		j.plaidBankAccounts, err = plaidClient.GetAccounts(ctx)
		if err != nil {
			log.ErrorContext(ctx, "failed to retrieve bank accounts from plaid", "err", err)
			return err
		}

		missingAccounts := j.findMissingAccounts(ctx)
		if len(missingAccounts) > 0 {
			log.InfoContext(ctx, "found newly inactive accounts, updating status", "count", len(missingAccounts))

			for _, bankAccount := range missingAccounts {
				bankAccount.Status = models.BankAccountStatusInactive
				j.log.DebugContext(ctx, "updating account to be inactive",
					"bankAccountId", bankAccount.BankAccountId,
					"plaid_bankAccountId", bankAccount.PlaidBankAccount.PlaidId,
				)
				if err := j.repo.UpdateBankAccount(ctx, &bankAccount); err != nil {
					log.ErrorContext(ctx, "failed to mark account as inactive", "err", err)
					continue
				}
			}
		} else {
			log.InfoContext(ctx, "no accounts to mark as inactive")
		}

		activeAccounts := j.findActiveAccounts(ctx)
		if len(activeAccounts) > 0 {
			log.InfoContext(ctx, "found reactivated accounts, updating status", "count", len(activeAccounts))

			for _, bankAccount := range activeAccounts {
				bankAccount.Status = models.BankAccountStatusActive
				j.log.DebugContext(ctx, "updating account to be reactivated",
					"bankAccountId", bankAccount.BankAccountId,
					"plaid_bankAccountId", bankAccount.PlaidBankAccount.PlaidId,
				)
				if err := j.repo.UpdateBankAccount(ctx, &bankAccount); err != nil {
					log.ErrorContext(ctx, "failed to mark account as active", "err", err)
					continue
				}
			}
		} else {
			log.InfoContext(ctx, "no accounts to mark as reactivated")
		}

		log.Log(ctx, logging.LevelTrace, "updating plaid link's last account sync timestamp")
		plaidLink.LastAccountSync = myownsanity.Pointer(ctx.Clock().Now())
		if err := j.repo.UpdatePlaidLink(ctx, plaidLink); err != nil {
			log.ErrorContext(ctx, "failed to update plaid link's last account sync timestamp", "err", err)
			return err
		}

		return nil
	})
}
