package background

import (
	"context"
	"fmt"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	RemoveLink = "RemoveLink"
)

var (
	_ JobHandler = &RemoveLinkHandler{}
	_ Job        = &RemoveLinkJob{}
)

type (
	RemoveLinkHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		publisher    pubsub.Publisher
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	RemoveLinkArguments struct {
		AccountId uint64 `json:"accountId"`
		LinkId    uint64 `json:"linkId"`
	}

	RemoveLinkJob struct {
		args      RemoveLinkArguments
		log       *logrus.Entry
		db        pg.DBI
		publisher pubsub.Publisher
		clock     clock.Clock
	}
)

// TriggerRemoveLink will dispatch a background job to remove the specified link and all of the data related to it from
// the desired account. This will return an error if the job fails to be enqueued, but does not indicate the status of
// the actual job.
func TriggerRemoveLink(ctx context.Context, backgroundJobs JobController, arguments RemoveLinkArguments) error {
	return backgroundJobs.EnqueueJob(ctx, RemoveLink, arguments)
}

func NewRemoveLinkHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	publisher pubsub.Publisher,
) *RemoveLinkHandler {
	return &RemoveLinkHandler{
		log:          log,
		db:           db,
		clock:        clock,
		publisher:    publisher,
		unmarshaller: DefaultJobUnmarshaller,
	}
}

func (r RemoveLinkHandler) QueueName() string {
	return RemoveLink
}

func (r *RemoveLinkHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args RemoveLinkArguments
	if err := errors.Wrap(r.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Remove Link job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return r.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		job, err := NewRemoveLinkJob(
			r.log.WithContext(span.Context()),
			txn,
			r.clock,
			r.publisher,
			args,
		)
		if err != nil {
			return err
		}

		return job.Run(span.Context())
	})
}

func NewRemoveLinkJob(
	log *logrus.Entry,
	db pg.DBI,
	clock clock.Clock,
	publisher pubsub.Publisher,
	args RemoveLinkArguments,
) (*RemoveLinkJob, error) {
	return &RemoveLinkJob{
		args:      args,
		log:       log,
		db:        db,
		publisher: publisher,
		clock:     clock,
	}, nil
}

func (r *RemoveLinkJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	accountId := r.args.AccountId
	linkId := r.args.LinkId

	repo := repository.NewRepositoryFromSession(r.clock, 0, accountId, r.db)

	log := r.log.WithContext(span.Context())

	link, err := repo.GetLink(span.Context(), linkId)
	if err != nil {
		crumbs.Warn(span.Context(), "failed to retrieve link to be removed, this job will not be retried", "weirdness", nil)
		log.WithError(err).Error("failed to retrieve link that to be removed, this job will not be retried")
		return nil
	}

	if link.PlaidLink != nil {
		crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.PlaidId)
	}

	bankAccountIds := make([]uint64, 0)
	{
		err = r.db.ModelContext(span.Context(), &models.BankAccount{}).
			Where(`"bank_account"."account_id" = ?`, accountId).
			Where(`"bank_account"."link_id" = ?`, linkId).
			Column("bank_account_id").
			Select(&bankAccountIds)
		if err != nil {
			log.WithError(err).Errorf("failed to retrieve bank account Ids for link")
			return errors.Wrap(err, "failed to retrieve bank account Ids for link")
		}
	}
	r.log = log.WithField("bankAccountIds", bankAccountIds)
	r.log.Info("removing data for bank account Ids for link")

	// Need to find these before we delete the transactions to avoid foreign key
	// issues and stray data being left behind.
	plaidTransactionIds := r.getPlaidTransactionsToRemove(span.Context(), bankAccountIds)
	plaidSyncIds := r.getPlaidSyncsToRemove(span.Context(), bankAccountIds)
	plaidBankAccountIds := r.getPlaidBankAccountsToRemove(span.Context(), bankAccountIds)
	plaidLinkIds := r.getPlaidLinksToRemove(span.Context())
	tellerTransactionIds := r.getTellerTransactionsToRemove(span.Context(), bankAccountIds)
	tellerSyncIds := r.getTellerSyncsToRemove(span.Context(), bankAccountIds)
	tellerBankAccountIds := r.getTellerBankAccountsToRemove(span.Context(), bankAccountIds)
	tellerLinkIds := r.getTellerLinksToRemove(span.Context())

	r.removeTransactionClusters(span.Context(), bankAccountIds)
	r.removeTransactions(span.Context(), bankAccountIds)
	r.removePlaidTransactions(span.Context(), plaidTransactionIds)
	r.removeTellerTransactions(span.Context(), tellerTransactionIds)
	r.removeSpending(span.Context(), bankAccountIds)
	r.removeFundingSchedules(span.Context(), bankAccountIds)
	r.removeBankAccounts(span.Context(), bankAccountIds)

	r.removePlaidSyncs(span.Context(), plaidSyncIds)
	r.removeTellerSyncs(span.Context(), tellerSyncIds)

	r.removePlaidBankAccounts(span.Context(), plaidBankAccountIds)
	r.removeTellerBankAccounts(span.Context(), tellerBankAccountIds)

	r.removeLink(span.Context())

	r.removePlaidLinks(span.Context(), plaidLinkIds)
	r.removeTellerLinks(span.Context(), tellerLinkIds)

	channelName := fmt.Sprintf("link:remove:%d:%d", accountId, linkId)
	if err = r.publisher.Notify(span.Context(), channelName, "success"); err != nil {
		log.WithError(err).Warn("failed to send notification about successfully removing link")
		crumbs.Warn(span.Context(), "failed to send notification about successfully removing link", "pubsub", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return nil
}

func (r *RemoveLinkJob) removeTransactionClusters(
	ctx context.Context,
	bankAccountIds []uint64,
) {
	result, err := r.db.ModelContext(ctx, &models.TransactionCluster{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove transaction clusters for link")
		panic(errors.Wrap(err, "failed to remove transaction clusters for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed transaction cluster(s)")
}

func (r *RemoveLinkJob) removeTransactions(
	ctx context.Context,
	bankAccountIds []uint64,
) {
	result, err := r.db.ModelContext(ctx, &models.Transaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove transactions for link")
		panic(errors.Wrap(err, "failed to remove transactions for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed transaction(s)")
}

func (r *RemoveLinkJob) getPlaidTransactionsToRemove(
	ctx context.Context,
	bankAccountIds []uint64,
) []uint64 {
	plaidTransactionIds := make([]uint64, 0)
	err := r.db.ModelContext(ctx, &models.PlaidTransaction{}).
		Join(`INNER JOIN "transactions" AS "transaction"`).
		JoinOn(`"plaid_transaction"."plaid_transaction_id" IN ("transaction"."plaid_transaction_id", "transaction"."pending_plaid_transaction_id")`).
		JoinOn(`"plaid_transaction"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.args.AccountId).
		WhereIn(`"transaction"."bank_account_id" IN (?)`, bankAccountIds).
		Column("plaid_transaction.plaid_transaction_id").
		Select(&plaidTransactionIds)
	if err != nil {
		panic(errors.Wrap(err, "failed to find plaid transactions to be removed"))
	}

	return plaidTransactionIds
}

func (r *RemoveLinkJob) removePlaidTransactions(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidTransaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_transaction_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid transactions for link")
		panic(errors.Wrap(err, "failed to remove plaid transactions for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid transaction(s)")
}

func (r *RemoveLinkJob) getTellerTransactionsToRemove(
	ctx context.Context,
	bankAccountIds []uint64,
) []uint64 {
	tellerTransactionIds := make([]uint64, 0)
	err := r.db.ModelContext(ctx, &models.TellerTransaction{}).
		Join(`INNER JOIN "transactions" AS "transaction"`).
		JoinOn(`"teller_transaction"."teller_transaction_id" = "transaction"."teller_transaction_id"`).
		JoinOn(`"teller_transaction"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.args.AccountId).
		WhereIn(`"transaction"."bank_account_id" IN (?)`, bankAccountIds).
		Column("teller_transaction.teller_transaction_id").
		Select(&tellerTransactionIds)
	if err != nil {
		panic(errors.Wrap(err, "failed to find teller transactions to remove"))
	}

	return tellerTransactionIds
}

func (r *RemoveLinkJob) removeTellerTransactions(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.TellerTransaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"teller_transaction_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove teller transactions for link")
		panic(errors.Wrap(err, "failed to remove teller transactions for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed teller transaction(s)")
}

func (r *RemoveLinkJob) removeSpending(
	ctx context.Context,
	bankAccountIds []uint64,
) {
	result, err := r.db.ModelContext(ctx, &models.Spending{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove spending for link")
		panic(errors.Wrap(err, "failed to remove spending for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed spending(s)")
}

func (r *RemoveLinkJob) removeFundingSchedules(
	ctx context.Context,
	bankAccountIds []uint64,
) {
	result, err := r.db.ModelContext(ctx, &models.FundingSchedule{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove funding schedules for link")
		panic(errors.Wrap(err, "failed to remove funding schedules for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed funding schedule(s)")
}

func (r *RemoveLinkJob) getPlaidSyncsToRemove(
	ctx context.Context,
	bankAccountIds []uint64,
) []uint64 {
	ids := make([]uint64, 0)
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

func (r *RemoveLinkJob) removePlaidSyncs(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidSync{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_sync_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid syncs for link")
		panic(errors.Wrap(err, "failed to remove plaid syncs for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid sync(s)")
}

func (r *RemoveLinkJob) getTellerSyncsToRemove(
	ctx context.Context,
	bankAccountIds []uint64,
) []uint64 {
	ids := make([]uint64, 0)
	err := r.db.ModelContext(ctx, &models.TellerSync{}).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"teller_sync"."teller_bank_account_id" = "bank_account"."teller_bank_account_id"`).
		JoinOn(`"teller_sync"."account_id" = "bank_account"."account_id"`).
		Where(`"teller_sync"."account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account"."bank_account_id" IN (?)`, bankAccountIds).
		Column("teller_sync_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find teller syncs to remove"))
	}

	return ids
}

func (r *RemoveLinkJob) removeTellerSyncs(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.TellerSync{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"teller_sync_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove teller syncs for link")
		panic(errors.Wrap(err, "failed to remove teller syncs for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed teller sync(s)")
}

func (r *RemoveLinkJob) getPlaidBankAccountsToRemove(
	ctx context.Context,
	bankAccountIds []uint64,
) []uint64 {
	ids := make([]uint64, 0)
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

func (r *RemoveLinkJob) removePlaidBankAccounts(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidBankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_bank_account_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid bank accounts for link")
		panic(errors.Wrap(err, "failed to remove plaid bank accounts for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid bank account(s)")
}

func (r *RemoveLinkJob) getTellerBankAccountsToRemove(
	ctx context.Context,
	bankAccountIds []uint64,
) []uint64 {
	ids := make([]uint64, 0)
	err := r.db.ModelContext(ctx, &models.TellerBankAccount{}).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"teller_bank_account"."teller_bank_account_id" = "bank_account"."teller_bank_account_id"`).
		JoinOn(`"teller_bank_account"."account_id" = "bank_account"."account_id"`).
		Where(`"teller_bank_account"."account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account"."bank_account_id" IN (?)`, bankAccountIds).
		Column("teller_bank_account.teller_bank_account_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find teller bank accounts to remove"))
	}

	return ids
}

func (r *RemoveLinkJob) removeTellerBankAccounts(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.TellerBankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"teller_bank_account_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove teller bank accounts for link")
		panic(errors.Wrap(err, "failed to remove teller bank accounts for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed teller bank account(s)")
}

func (r *RemoveLinkJob) getPlaidLinksToRemove(
	ctx context.Context,
) []uint64 {
	ids := make([]uint64, 0)
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

func (r *RemoveLinkJob) removePlaidLinks(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.PlaidLink{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_link_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid links for link")
		panic(errors.Wrap(err, "failed to remove plaid links for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid link(s)")
}

func (r *RemoveLinkJob) getTellerLinksToRemove(
	ctx context.Context,
) []uint64 {
	ids := make([]uint64, 0)
	err := r.db.ModelContext(ctx, &models.TellerLink{}).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"teller_link"."teller_link_id" = "link"."teller_link_id"`).
		JoinOn(`"teller_link"."account_id" = "link"."account_id"`).
		Where(`"teller_link"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."link_id" = ?`, r.args.LinkId).
		Column("teller_link.teller_link_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find teller links to remove"))
	}

	return ids
}

func (r *RemoveLinkJob) removeTellerLinks(
	ctx context.Context,
	ids []uint64,
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &models.TellerLink{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"teller_link_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove teller links for link")
		panic(errors.Wrap(err, "failed to remove teller links for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed teller link(s)")
}

func (r *RemoveLinkJob) removeBankAccounts(
	ctx context.Context,
	bankAccountIds []uint64,
) {
	result, err := r.db.ModelContext(ctx, &models.BankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove bank accounts for link")
		panic(errors.Wrap(err, "failed to remove bank accounts for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed bank account(s)")
}

func (r *RemoveLinkJob) removeLink(
	ctx context.Context,
) {
	result, err := r.db.ModelContext(ctx, &models.Link{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		Where(`"link_id" = ?`, r.args.LinkId).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove link")
		panic(errors.Wrap(err, "failed to remove link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed link")
}
