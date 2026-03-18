package link_jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type RemoveLinkArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
	LinkId    models.ID[models.Link]    `json:"linkId"`
}

type removeLinkJob struct {
	args      RemoveLinkArguments
	log       *slog.Logger
	db        pg.DBI
	publisher pubsub.Publisher
	clock     clock.Clock
}

func RemoveLink(ctx queue.Context, args RemoveLinkArguments) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"linkId", args.LinkId,
		)
		accountId := args.AccountId
		linkId := args.LinkId

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_system",
			accountId,
			ctx.DB(),
			log,
		)

		link, err := repo.GetLink(ctx, linkId)
		if err != nil {
			crumbs.Warn(ctx, "failed to retrieve link to be removed, this job will not be retried", "weirdness", nil)
			log.ErrorContext(ctx, "failed to retrieve link that to be removed, this job will not be retried", "err", err)
			return nil
		}

		if link.PlaidLink != nil {
			crumbs.IncludePlaidItemIDTag(
				sentry.SpanFromContext(ctx),
				link.PlaidLink.PlaidId,
			)
		}

		bankAccountIds := make([]models.ID[models.BankAccount], 0)
		{
			err = ctx.DB().ModelContext(ctx, &models.BankAccount{}).
				Where(`"bank_account"."account_id" = ?`, accountId).
				Where(`"bank_account"."link_id" = ?`, linkId).
				Column("bank_account_id").
				Select(&bankAccountIds)
			if err != nil {
				log.ErrorContext(ctx, "failed to retrieve bank account Ids for link", "err", err)
				return errors.Wrap(err, "failed to retrieve bank account Ids for link")
			}
		}

		r := removeLinkJob{
			args:      args,
			log:       log.With("bankAccountIds", bankAccountIds),
			db:        ctx.DB(),
			publisher: ctx.Publisher(),
			clock:     ctx.Clock(),
		}

		r.log.InfoContext(ctx, "removing data for bank account Ids for link")

		secretIds := r.getSecretsToRemove(ctx)

		// Need to find these before we delete the transactions to avoid foreign key
		// issues and stray data being left behind.
		plaidTransactionIds := r.getPlaidTransactionsToRemove(ctx, bankAccountIds)
		plaidSyncIds := r.getPlaidSyncsToRemove(ctx, bankAccountIds)
		plaidBankAccountIds := r.getPlaidBankAccountsToRemove(ctx, bankAccountIds)
		plaidLinkIds := r.getPlaidLinksToRemove(ctx)

		lunchFlowTransactionIds := r.getLunchFlowTransactionsToRemove(ctx)
		lunchFlowBankAccountIds := r.getLunchFlowBankAccountsToRemove(ctx)
		lunchFlowLinkIds := r.getLunchFlowLinksToRemove(ctx)

		r.removeTransactionClusters(ctx, bankAccountIds)
		// TODO Also remove any non-reconciled files. Really this should also be
		// calling to files in S3 or the underlying to remove them.
		r.removeTransactionUploads(ctx, bankAccountIds)
		r.removeTransactions(ctx, bankAccountIds)
		r.removePlaidTransactions(ctx, plaidTransactionIds)
		r.removeLunchFlowTransactions(ctx, lunchFlowTransactionIds)
		r.removeSpending(ctx, bankAccountIds)
		r.removeFundingSchedules(ctx, bankAccountIds)
		r.removeBankAccounts(ctx, bankAccountIds)
		r.removePlaidSyncs(ctx, plaidSyncIds)
		r.removePlaidBankAccounts(ctx, plaidBankAccountIds)
		r.removeLunchFlowBankAccounts(ctx, lunchFlowBankAccountIds)
		r.removeLink(ctx)
		r.removePlaidLinks(ctx, plaidLinkIds)
		r.removeLunchFlowLinks(ctx, lunchFlowLinkIds)
		r.removeSecrets(ctx, secretIds)

		channelName := fmt.Sprintf("link:remove:%s:%s", accountId, linkId)
		if err = r.publisher.Notify(
			ctx,
			accountId,
			channelName,
			"success",
		); err != nil {
			log.WarnContext(
				ctx,
				"failed to send notification about successfully removing link",
				"err", err,
			)
			crumbs.Warn(
				ctx,
				"failed to send notification about successfully removing link",
				"pubsub",
				map[string]any{
					"error": err.Error(),
				},
			)
		}

		return nil
	})
}

func (r *removeLinkJob) removeTransactionClusters(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) {
	if len(bankAccountIds) == 0 {
		return
	}
	result, err := r.db.ModelContext(ctx, &models.TransactionCluster{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove transaction clusters for link", "err", err)
		panic(errors.Wrap(err, "failed to remove transaction clusters for link"))
	}

	r.log.InfoContext(ctx, "removed transaction cluster(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) removeTransactionUploads(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) {
	if len(bankAccountIds) == 0 {
		return
	}
	result, err := r.db.ModelContext(ctx, &models.TransactionUpload{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove transaction uploads for link", "err", err)
		panic(errors.Wrap(err, "failed to remove transaction uploads for link"))
	}

	r.log.InfoContext(ctx, "removed transaction upload(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) removeTransactions(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) {
	if len(bankAccountIds) == 0 {
		return
	}
	result, err := r.db.ModelContext(ctx, &models.Transaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove transactions for link", "err", err)
		panic(errors.Wrap(err, "failed to remove transactions for link"))
	}

	r.log.InfoContext(ctx, "removed transaction(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getLunchFlowTransactionsToRemove(
	ctx context.Context,
) []models.ID[models.LunchFlowTransaction] {
	ids := make([]models.ID[models.LunchFlowTransaction], 0)
	err := r.db.ModelContext(ctx, &models.LunchFlowTransaction{}).
		Join(`INNER JOIN "lunch_flow_bank_accounts" AS "lunch_flow_bank_account"`).
		JoinOn(`"lunch_flow_bank_account"."lunch_flow_bank_account_id" = "lunch_flow_transaction"."lunch_flow_bank_account_id"`).
		JoinOn(`"lunch_flow_bank_account"."account_id" = "lunch_flow_transaction"."account_id"`).
		// Lunch Flow bank accounts might not be associated with a monetr bank
		// account if they are inactive, so we instead need to traverse upwards to
		// the link instead.
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"link"."lunch_flow_link_id" = "lunch_flow_bank_account"."lunch_flow_link_id"`).
		JoinOn(`"link"."account_id" = "lunch_flow_bank_account"."account_id"`).
		Where(`"lunch_flow_transaction"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."link_id" = ?`, r.args.LinkId).
		Column("lunch_flow_transaction.lunch_flow_transaction_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find Lunch Flow transactions to be removed"))
	}

	return ids
}

func (r *removeLinkJob) removeLunchFlowTransactions(
	ctx context.Context,
	ids []models.ID[models.LunchFlowTransaction],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.LunchFlowTransaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"lunch_flow_transaction_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove Lunch Flow transactions for link", "err", err)
		panic(errors.Wrap(err, "failed to remove Lunch Flow transactions for link"))
	}

	r.log.InfoContext(ctx, "removed Lunch Flow transaction(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getPlaidTransactionsToRemove(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) []models.ID[models.PlaidTransaction] {
	if len(bankAccountIds) == 0 {
		return nil
	}
	plaidTransactionIds := make([]models.ID[models.PlaidTransaction], 0)
	err := r.db.ModelContext(ctx, &models.PlaidTransaction{}).
		Join(`INNER JOIN "plaid_bank_accounts" AS "plaid_bank_account"`).
		JoinOn(`"plaid_bank_account"."plaid_bank_account_id" = "plaid_transaction"."plaid_bank_account_id"`).
		JoinOn(`"plaid_bank_account"."account_id" = "plaid_transaction"."account_id"`).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."plaid_bank_account_id" = "plaid_bank_account"."plaid_bank_account_id"`).
		JoinOn(`"bank_account"."account_id" = "plaid_bank_account"."account_id"`).
		Where(`"plaid_transaction"."account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account"."bank_account_id" IN (?)`, bankAccountIds).
		Column("plaid_transaction.plaid_transaction_id").
		Select(&plaidTransactionIds)
	if err != nil {
		panic(errors.Wrap(err, "failed to find plaid transactions to be removed"))
	}

	return plaidTransactionIds
}

func (r *removeLinkJob) removePlaidTransactions(
	ctx context.Context,
	ids []models.ID[models.PlaidTransaction],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidTransaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_transaction_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove plaid transactions for link", "err", err)
		panic(errors.Wrap(err, "failed to remove plaid transactions for link"))
	}

	r.log.InfoContext(ctx, "removed plaid transaction(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) removeSpending(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) {
	if len(bankAccountIds) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.Spending{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove spending for link", "err", err)
		panic(errors.Wrap(err, "failed to remove spending for link"))
	}

	r.log.InfoContext(ctx, "removed spending(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) removeFundingSchedules(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) {
	if len(bankAccountIds) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.FundingSchedule{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove funding schedules for link", "err", err)
		panic(errors.Wrap(err, "failed to remove funding schedules for link"))
	}

	r.log.InfoContext(ctx, "removed funding schedule(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getPlaidSyncsToRemove(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) []models.ID[models.PlaidSync] {
	if len(bankAccountIds) == 0 {
		return nil
	}
	ids := make([]models.ID[models.PlaidSync], 0)
	err := r.db.ModelContext(ctx, &models.PlaidSync{}).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"plaid_sync"."plaid_link_id" = "link"."plaid_link_id"`).
		JoinOn(`"plaid_sync"."account_id" = "link"."account_id"`).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"link"."link_id" = "bank_account"."link_id"`).
		JoinOn(`"link"."account_id" = "bank_account"."account_id"`).
		Where(`"plaid_sync"."account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account"."bank_account_id" IN (?)`, bankAccountIds).
		Column("plaid_sync_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find plaid syncs to remove"))
	}

	return ids
}

func (r *removeLinkJob) removePlaidSyncs(
	ctx context.Context,
	ids []models.ID[models.PlaidSync],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidSync{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_sync_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove plaid syncs for link", "err", err)
		panic(errors.Wrap(err, "failed to remove plaid syncs for link"))
	}

	r.log.InfoContext(ctx, "removed plaid sync(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getLunchFlowBankAccountsToRemove(
	ctx context.Context,
) []models.ID[models.LunchFlowBankAccount] {
	ids := make([]models.ID[models.LunchFlowBankAccount], 0)
	err := r.db.ModelContext(ctx, &models.LunchFlowBankAccount{}).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"link"."lunch_flow_link_id" = "lunch_flow_bank_account"."lunch_flow_link_id"`).
		JoinOn(`"link"."account_id" = "lunch_flow_bank_account"."account_id"`).
		Where(`"lunch_flow_bank_account"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."link_id" = ?`, r.args.LinkId).
		Column("lunch_flow_bank_account.lunch_flow_bank_account_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find lunch_flow bank accounts to remove"))
	}

	return ids
}

func (r *removeLinkJob) removeLunchFlowBankAccounts(
	ctx context.Context,
	ids []models.ID[models.LunchFlowBankAccount],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.LunchFlowBankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"lunch_flow_bank_account_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove Lunch Flow bank accounts for link", "err", err)
		panic(errors.Wrap(err, "failed to remove Lunch Flow bank accounts for link"))
	}

	r.log.InfoContext(ctx, "removed Lunch Flow bank account(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getPlaidBankAccountsToRemove(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) []models.ID[models.PlaidBankAccount] {
	if len(bankAccountIds) == 0 {
		return nil
	}
	ids := make([]models.ID[models.PlaidBankAccount], 0)
	err := r.db.ModelContext(ctx, &models.PlaidBankAccount{}).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"plaid_bank_account"."plaid_bank_account_id" = "bank_account"."plaid_bank_account_id"`).
		JoinOn(`"plaid_bank_account"."account_id" = "bank_account"."account_id"`).
		Where(`"plaid_bank_account"."account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account"."bank_account_id" IN (?)`, bankAccountIds).
		Column("plaid_bank_account.plaid_bank_account_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find plaid bank accounts to remove"))
	}

	return ids
}

func (r *removeLinkJob) removePlaidBankAccounts(
	ctx context.Context,
	ids []models.ID[models.PlaidBankAccount],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidBankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_bank_account_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove plaid bank accounts for link", "err", err)
		panic(errors.Wrap(err, "failed to remove plaid bank accounts for link"))
	}

	r.log.InfoContext(ctx, "removed plaid bank account(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getPlaidLinksToRemove(
	ctx context.Context,
) []models.ID[models.PlaidLink] {
	ids := make([]models.ID[models.PlaidLink], 0)
	err := r.db.ModelContext(ctx, &models.PlaidLink{}).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"plaid_link"."plaid_link_id" = "link"."plaid_link_id"`).
		JoinOn(`"plaid_link"."account_id" = "link"."account_id"`).
		Where(`"plaid_link"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."link_id" = ?`, r.args.LinkId).
		Column("plaid_link.plaid_link_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find plaid links to remove"))
	}

	return ids
}

func (r *removeLinkJob) removePlaidLinks(
	ctx context.Context,
	ids []models.ID[models.PlaidLink],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidLink{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_link_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove plaid links for link", "err", err)
		panic(errors.Wrap(err, "failed to remove plaid links for link"))
	}

	r.log.InfoContext(ctx, "removed plaid link(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getLunchFlowLinksToRemove(
	ctx context.Context,
) []models.ID[models.LunchFlowLink] {
	ids := make([]models.ID[models.LunchFlowLink], 0)
	err := r.db.ModelContext(ctx, &models.LunchFlowLink{}).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"lunch_flow_link"."lunch_flow_link_id" = "link"."lunch_flow_link_id"`).
		JoinOn(`"lunch_flow_link"."account_id" = "link"."account_id"`).
		Where(`"lunch_flow_link"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."link_id" = ?`, r.args.LinkId).
		Column("lunch_flow_link.lunch_flow_link_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find Lunch Flow links to remove"))
	}

	return ids
}

func (r *removeLinkJob) removeLunchFlowLinks(
	ctx context.Context,
	ids []models.ID[models.LunchFlowLink],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.LunchFlowLink{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"lunch_flow_link_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove Lunch Flow links for link", "err", err)
		panic(errors.Wrap(err, "failed to remove Lunch Flow links for link"))
	}

	r.log.InfoContext(ctx, "removed Lunch Flow link(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) getSecretsToRemove(
	ctx context.Context,
) []models.ID[models.Secret] {
	ids := make([]models.ID[models.Secret], 0)
	err := r.db.ModelContext(ctx, &models.Secret{}).
		// The secret can be associated with either a Plaid link or a Lunch Flow
		// link. But must be associated with the desired link from the arguments!
		Join(`LEFT JOIN "plaid_links" as "plaid_link"`).
		JoinOn(`"plaid_link"."secret_id" = "secret"."secret_id"`).
		JoinOn(`"plaid_link"."account_id" = "secret"."account_id"`).
		Join(`LEFT JOIN "lunch_flow_links" as "lunch_flow_link"`).
		JoinOn(`"lunch_flow_link"."secret_id" = "secret"."secret_id"`).
		JoinOn(`"lunch_flow_link"."account_id" = "secret"."account_id"`).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`("plaid_link"."plaid_link_id" = "link"."plaid_link_id" AND "plaid_link"."account_id" = "link"."account_id") OR ("lunch_flow_link"."lunch_flow_link_id" = "link"."lunch_flow_link_id" AND "lunch_flow_link"."account_id" = "link"."account_id")`).
		Where(`"link"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."link_id" = ?`, r.args.LinkId).
		Column("secret.secret_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find secrets to remove"))
	}

	return ids
}

func (r *removeLinkJob) removeSecrets(
	ctx context.Context,
	ids []models.ID[models.Secret],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.Secret{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"secret_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove secrets for link", "err", err)
		panic(errors.Wrap(err, "failed to remove secrets for link"))
	}

	r.log.InfoContext(ctx, "removed secret(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) removeBankAccounts(
	ctx context.Context,
	bankAccountIds []models.ID[models.BankAccount],
) {
	if len(bankAccountIds) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.BankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove bank accounts for link", "err", err)
		panic(errors.Wrap(err, "failed to remove bank accounts for link"))
	}

	r.log.InfoContext(ctx, "removed bank account(s)", "removed", result.RowsAffected())
}

func (r *removeLinkJob) removeLink(
	ctx context.Context,
) {
	result, err := r.db.ModelContext(ctx, &models.Link{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		Where(`"link_id" = ?`, r.args.LinkId).
		Delete()
	if err != nil {
		r.log.ErrorContext(ctx, "failed to remove link", "err", err)
		panic(errors.Wrap(err, "failed to remove link"))
	}

	r.log.InfoContext(ctx, "removed link", "removed", result.RowsAffected())
}
