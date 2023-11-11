package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	RemoveTransactions = "RemoveTransactions"
)

var (
	_ JobHandler = &RemoveTransactionsHandler{}
	_ Job        = &RemoveTransactionsJob{}
)

type (
	RemoveTransactionsHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	RemoveTransactionsArguments struct {
		AccountId           uint64   `json:"accountId"`
		LinkId              uint64   `json:"linkId"`
		PlaidTransactionIds []string `json:"plaidTransactionIds"`
	}

	RemoveTransactionsJob struct {
		args  RemoveTransactionsArguments
		log   *logrus.Entry
		repo  repository.BaseRepository
		clock clock.Clock
	}
)

func TriggerRemoveTransactions(ctx context.Context, backgroundJobs JobController, arguments RemoveTransactionsArguments) error {
	return backgroundJobs.TriggerJob(ctx, RemoveTransactions, arguments)
}

func NewRemoveTransactionsHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
) *RemoveTransactionsHandler {
	return &RemoveTransactionsHandler{
		log:          log,
		db:           db,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (r RemoveTransactionsHandler) QueueName() string {
	return RemoveTransactions
}

func (r *RemoveTransactionsHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args RemoveTransactionsArguments
	if err := errors.Wrap(r.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Remove Transactions job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)
	crumbs.Debug(ctx, "Removing transactions", map[string]interface{}{
		"linkId":              args.LinkId,
		"plaidTransactionIds": args.PlaidTransactionIds,
	})

	return r.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		repo := repository.NewRepositoryFromSession(r.clock, 0, args.AccountId, txn)
		job, err := NewRemoveTransactionsJob(r.log.WithContext(ctx), repo, r.clock, args)
		if err != nil {
			return err
		}
		return job.Run(span.Context())
	})
}

func NewRemoveTransactionsJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	args RemoveTransactionsArguments,
) (*RemoveTransactionsJob, error) {
	return &RemoveTransactionsJob{
		args:  args,
		log:   log,
		repo:  repo,
		clock: clock,
	}, nil
}

func (r *RemoveTransactionsJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := r.log.WithContext(span.Context())

	log.Infof("removing %d transaction(s)", len(r.args.PlaidTransactionIds))

	link, err := r.repo.GetLink(span.Context(), r.args.LinkId)
	if err != nil {
		log.WithError(err).Error("failed to retrieve link details to pull transactions")
		return err
	}

	log = log.WithField("linkId", link.LinkId)

	if link.PlaidLink == nil {
		err = errors.Errorf("cannot pull account balanaces for link without plaid info")
		log.WithError(err).Errorf("failed to pull transactions")
		return err
	}

	transactions, err := r.repo.GetTransactionsByPlaidTransactionId(span.Context(), r.args.LinkId, r.args.PlaidTransactionIds)
	if err != nil {
		log.WithError(err).Error("failed to retrieve transactions by plaid transaction Id for removal")
		return err
	}

	if len(transactions) == 0 {
		log.Warnf("no transactions retrieved, nothing to be done. transactions might already have been deleted")
		return nil
	}

	if len(transactions) != len(r.args.PlaidTransactionIds) {
		log.Warnf("number of transactions retrieved does not match expected number of transactions, expected: %d found: %d", len(r.args.PlaidTransactionIds), len(transactions))
		crumbs.IndicateBug(span.Context(), "The number of transactions retrieved does not match the expected number of transactions", map[string]interface{}{
			"expected":            len(r.args.PlaidTransactionIds),
			"found":               len(transactions),
			"plaidTransactionIds": r.args.PlaidTransactionIds,
		})
	}

	for _, existingTransaction := range transactions {
		if existingTransaction.SpendingId == nil {
			continue
		}

		// If the transaction is spent from something then we need to remove the spent from before deleting it to
		// maintain our balances correctly.
		updatedTransaction := existingTransaction
		updatedTransaction.SpendingId = nil

		// This is a simple sanity check, working with objects in slices and for loops can be goofy, or my
		// understanding of the way objects works with how they are referenced in memory is poor. This is to make
		// sure im not doing it wrong though. I'm worried that making a "copy" of the object and then modifying the
		// copy will modify the original as well.
		if existingTransaction.SpendingId == nil {
			sentry.CaptureMessage("original transaction modified")
			panic("original transaction modified")
		}

		_, err = r.repo.ProcessTransactionSpentFrom(
			span.Context(),
			existingTransaction.BankAccountId,
			&updatedTransaction,
			&existingTransaction,
		)
		if err != nil {
			return err
		}
	}

	for _, transaction := range transactions {
		if err = r.repo.DeleteTransaction(span.Context(), transaction.BankAccountId, transaction.TransactionId); err != nil {
			log.WithField("transactionId", transaction.TransactionId).WithError(err).
				Error("failed to delete transaction")
			return err
		}
	}

	log.Debugf("successfully removed %d transaction(s)", len(transactions))

	link.LastSuccessfulUpdate = myownsanity.TimeP(r.clock.Now().UTC())
	return r.repo.UpdateLink(span.Context(), link)
}
