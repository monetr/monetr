package background

import (
	"context"
	"fmt"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	RemoveLink = "RemoveLink"
)

var (
	_ JobHandler        = &RemoveLinkHandler{}
	_ JobImplementation = &RemoveLinkJob{}
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
		AccountId ID[Account] `json:"accountId"`
		LinkId    ID[Link]    `json:"linkId"`
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

func (r *RemoveLinkHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args RemoveLinkArguments
	if err := errors.Wrap(r.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Remove Link job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return r.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context()).WithFields(logrus.Fields{
			"accountId": args.AccountId,
			"linkId":    args.LinkId,
		})

		job, err := NewRemoveLinkJob(
			log,
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

	log := r.log.WithContext(span.Context())

	accountId := r.args.AccountId
	linkId := r.args.LinkId

	repo := repository.NewRepositoryFromSession(
		r.clock,
		"user_system",
		accountId,
		r.db,
		log,
	)

	link, err := repo.GetLink(span.Context(), linkId)
	if err != nil {
		crumbs.Warn(span.Context(), "failed to retrieve link to be removed, this job will not be retried", "weirdness", nil)
		log.WithError(err).Error("failed to retrieve link that to be removed, this job will not be retried")
		return nil
	}

	if link.PlaidLink != nil {
		crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.PlaidId)
	}

	bankAccountIds := make([]ID[BankAccount], 0)
	{
		err = r.db.ModelContext(span.Context(), &BankAccount{}).
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

	lunchFlowTransactionIds := r.getLunchFlowTransactionsToRemove(span.Context())
	lunchFlowBankAccountIds := r.getLunchFlowBankAccountsToRemove(span.Context())
	// TODO Also read lunch flow links

	r.removeTransactionClusters(span.Context(), bankAccountIds)
	// TODO Also remove any non-reconciled files
	r.removeTransactionUploads(span.Context(), bankAccountIds)
	r.removeTransactions(span.Context(), bankAccountIds)
	r.removePlaidTransactions(span.Context(), plaidTransactionIds)
	r.removeLunchFlowTransactions(span.Context(), lunchFlowTransactionIds)
	r.removeSpending(span.Context(), bankAccountIds)
	r.removeFundingSchedules(span.Context(), bankAccountIds)
	r.removeBankAccounts(span.Context(), bankAccountIds)
	r.removePlaidSyncs(span.Context(), plaidSyncIds)
	r.removePlaidBankAccounts(span.Context(), plaidBankAccountIds)
	r.removeLunchFlowBankAccounts(span.Context(), lunchFlowBankAccountIds)
	r.removeLink(span.Context())
	r.removePlaidLinks(span.Context(), plaidLinkIds)

	channelName := fmt.Sprintf("link:remove:%s:%s", accountId, linkId)
	if err = r.publisher.Notify(span.Context(), channelName, "success"); err != nil {
		log.WithError(err).Warn("failed to send notification about successfully removing link")
		crumbs.Warn(span.Context(), "failed to send notification about successfully removing link", "pubsub", map[string]any{
			"error": err.Error(),
		})
	}

	return nil
}

func (r *RemoveLinkJob) removeTransactionClusters(
	ctx context.Context,
	bankAccountIds []ID[BankAccount],
) {
	result, err := r.db.ModelContext(ctx, &TransactionCluster{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove transaction clusters for link")
		panic(errors.Wrap(err, "failed to remove transaction clusters for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed transaction cluster(s)")
}

func (r *RemoveLinkJob) removeTransactionUploads(
	ctx context.Context,
	bankAccountIds []ID[BankAccount],
) {
	result, err := r.db.ModelContext(ctx, &TransactionUpload{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove transaction uploads for link")
		panic(errors.Wrap(err, "failed to remove transaction uploads for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed transaction upload(s)")
}

func (r *RemoveLinkJob) removeTransactions(
	ctx context.Context,
	bankAccountIds []ID[BankAccount],
) {
	result, err := r.db.ModelContext(ctx, &Transaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"bank_account_id" IN (?)`, bankAccountIds).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove transactions for link")
		panic(errors.Wrap(err, "failed to remove transactions for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed transaction(s)")
}

func (r *RemoveLinkJob) getLunchFlowTransactionsToRemove(
	ctx context.Context,
) []ID[LunchFlowTransaction] {
	ids := make([]ID[LunchFlowTransaction], 0)
	err := r.db.ModelContext(ctx, &LunchFlowTransaction{}).
		Join(`INNER JOIN "lunch_flow_bank_accounts" AS "lunch_flow_bank_account"`).
		JoinOn(`"lunch_flow_bank_account"."lunch_flow_bank_account_id" = "lunch_flow_transaction"."lunch_flow_bank_account_id"`).
		JoinOn(`"lunch_flow_bank_account"."account_id" = "lunch_flow_transaction"."account_id"`).
		// Lunch flow bank accounts might not be associated with a monetr bank
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
		panic(errors.Wrap(err, "failed to find lunch flow transactions to be removed"))
	}

	return ids
}

func (r *RemoveLinkJob) removeLunchFlowTransactions(
	ctx context.Context,
	ids []ID[LunchFlowTransaction],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &LunchFlowTransaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"lunch_flow_transaction_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove lunch flow transactions for link")
		panic(errors.Wrap(err, "failed to remove lunch flow transactions for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed lunch flow transaction(s)")
}

func (r *RemoveLinkJob) getPlaidTransactionsToRemove(
	ctx context.Context,
	bankAccountIds []ID[BankAccount],
) []ID[PlaidTransaction] {
	plaidTransactionIds := make([]ID[PlaidTransaction], 0)
	err := r.db.ModelContext(ctx, &PlaidTransaction{}).
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

func (r *RemoveLinkJob) removePlaidTransactions(
	ctx context.Context,
	ids []ID[PlaidTransaction],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &PlaidTransaction{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_transaction_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid transactions for link")
		panic(errors.Wrap(err, "failed to remove plaid transactions for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid transaction(s)")
}

func (r *RemoveLinkJob) removeSpending(
	ctx context.Context,
	bankAccountIds []ID[BankAccount],
) {
	result, err := r.db.ModelContext(ctx, &Spending{}).
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
	bankAccountIds []ID[BankAccount],
) {
	result, err := r.db.ModelContext(ctx, &FundingSchedule{}).
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
	bankAccountIds []ID[BankAccount],
) []ID[PlaidSync] {
	ids := make([]ID[PlaidSync], 0)
	err := r.db.ModelContext(ctx, &PlaidSync{}).
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
	ids []ID[PlaidSync],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &PlaidSync{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_sync_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid syncs for link")
		panic(errors.Wrap(err, "failed to remove plaid syncs for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid sync(s)")
}

func (r *RemoveLinkJob) getLunchFlowBankAccountsToRemove(
	ctx context.Context,
) []ID[LunchFlowBankAccount] {
	ids := make([]ID[LunchFlowBankAccount], 0)
	err := r.db.ModelContext(ctx, &LunchFlowBankAccount{}).
		Join(`INNER JOIN "links" AS "link"`).
		JoinOn(`"link"."lunch_flow_link_id" = "lunch_flow_bank_account"."lunch_flow_link_id"`).
		JoinOn(`"link"."account_id" = "lunch_flow_bank_account"."account_id"`).
		Where(`"lunch_flow_bank_account"."account_id" = ?`, r.args.AccountId).
		Where(`"link"."account_id" = ?`, r.args.AccountId).
		Column("lunch_flow_bank_account.lunch_flow_bank_account_id").
		Select(&ids)
	if err != nil {
		panic(errors.Wrap(err, "failed to find lunch_flow bank accounts to remove"))
	}

	return ids
}

func (r *RemoveLinkJob) removeLunchFlowBankAccounts(
	ctx context.Context,
	ids []ID[LunchFlowBankAccount],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &LunchFlowBankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"lunch_flow_bank_account_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove lunch flow bank accounts for link")
		panic(errors.Wrap(err, "failed to remove lunch flow bank accounts for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed lunch flow bank account(s)")
}

func (r *RemoveLinkJob) getPlaidBankAccountsToRemove(
	ctx context.Context,
	bankAccountIds []ID[BankAccount],
) []ID[PlaidBankAccount] {
	ids := make([]ID[PlaidBankAccount], 0)
	err := r.db.ModelContext(ctx, &PlaidBankAccount{}).
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
	ids []ID[PlaidBankAccount],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &PlaidBankAccount{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_bank_account_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid bank accounts for link")
		panic(errors.Wrap(err, "failed to remove plaid bank accounts for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid bank account(s)")
}

func (r *RemoveLinkJob) getPlaidLinksToRemove(
	ctx context.Context,
) []ID[PlaidLink] {
	ids := make([]ID[PlaidLink], 0)
	err := r.db.ModelContext(ctx, &PlaidLink{}).
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
	ids []ID[PlaidLink],
) {
	if len(ids) == 0 {
		return
	}

	result, err := r.db.ModelContext(ctx, &PlaidLink{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		WhereIn(`"plaid_link_id" IN (?)`, ids).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove plaid links for link")
		panic(errors.Wrap(err, "failed to remove plaid links for link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed plaid link(s)")
}

func (r *RemoveLinkJob) removeBankAccounts(
	ctx context.Context,
	bankAccountIds []ID[BankAccount],
) {
	result, err := r.db.ModelContext(ctx, &BankAccount{}).
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
	result, err := r.db.ModelContext(ctx, &Link{}).
		Where(`"account_id" = ?`, r.args.AccountId).
		Where(`"link_id" = ?`, r.args.LinkId).
		Delete()
	if err != nil {
		r.log.WithError(err).Errorf("failed to remove link")
		panic(errors.Wrap(err, "failed to remove link"))
	}

	r.log.WithField("removed", result.RowsAffected()).Info("removed link")
}
