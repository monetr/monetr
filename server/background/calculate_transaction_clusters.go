package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/recurring"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	CalculateTransactionClusters = "CalculateTransactionClusters"
)

var (
	_ JobHandler        = &CalculateTransactionClustersHandler{}
	_ JobImplementation = &CalculateTransactionClustersJob{}
)

type CalculateTransactionClustersHandler struct {
	log          *logrus.Entry
	db           pg.DBI
	clock        clock.Clock
	unmarshaller JobUnmarshaller
}

type CalculateTransactionClustersArguments struct {
	AccountId     ID[Account]     `json:"accountId"`
	BankAccountId ID[BankAccount] `json:"bankAccountId"`
}

type CalculateTransactionClustersJob struct {
	args  CalculateTransactionClustersArguments
	log   *logrus.Entry
	db    pg.DBI
	clock clock.Clock
}

func TriggerCalculateTransactionClusters(
	ctx context.Context,
	backgroundJobs BackgroundJobs,
	arguments CalculateTransactionClustersArguments,
) error {
	return backgroundJobs.EnqueueJob(ctx, CalculateTransactionClusters, arguments)
}

func NewCalculateTransactionClustersHandler(
	log *logrus.Entry,
	db pg.DBI,
	clock clock.Clock,
) *CalculateTransactionClustersHandler {
	return &CalculateTransactionClustersHandler{
		log:          log,
		db:           db,
		clock:        clock,
		unmarshaller: DefaultJobUnmarshaller,
	}
}

func (c CalculateTransactionClustersHandler) QueueName() string {
	return CalculateTransactionClusters
}

func (c *CalculateTransactionClustersHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	var args CalculateTransactionClustersArguments
	if err := errors.Wrap(c.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Calculate Transaction Clusters job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return c.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		job, err := NewCalculateTransactionClustersJob(
			log.WithContext(span.Context()),
			txn,
			c.clock,
			args,
		)
		if err != nil {
			return err
		}

		return job.Run(span.Context())
	})
}

func NewCalculateTransactionClustersJob(
	log *logrus.Entry,
	db pg.DBI,
	clock clock.Clock,
	args CalculateTransactionClustersArguments,
) (*CalculateTransactionClustersJob, error) {
	return &CalculateTransactionClustersJob{
		args:  args,
		log:   log,
		db:    db,
		clock: clock,
	}, nil
}

func (c *CalculateTransactionClustersJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	accountId := c.args.AccountId
	bankAccountId := c.args.BankAccountId

	repo := repository.NewRepositoryFromSession(c.clock, "user_system", accountId, c.db)

	log := c.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"accountId":     accountId,
		"bankAccountId": bankAccountId,
	})

	clustering := recurring.NewSimilarTransactions_TFIDF_DBSCAN(log)

	limit := 500
	offset := 0
	for {
		txnLog := log.WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		})
		txnLog.Trace("requesting next batch of transactions")
		transactions, err := repo.GetTransactions(span.Context(), bankAccountId, limit, offset)
		if err != nil {
			return errors.Wrap(err, "failed to read transactions for clustering")
		}
		txnLog = log.WithField("count", len(transactions))

		for i := range transactions {
			clustering.AddTransaction(&transactions[i])
		}

		if len(transactions) < limit {
			txnLog.Trace("reached end of transactions")
			break
		}

		offset += len(transactions)
	}

	result := clustering.DetectSimilarTransactions(span.Context())

	if len(result) == 0 {
		log.Info("no similar transactions detected, nothing to persist")
		return nil
	}

	log.WithFields(logrus.Fields{
		"clusters": len(result),
	}).Info("similar transaction clusters detected")

	if err := repo.WriteTransactionClusters(span.Context(), bankAccountId, result); err != nil {
		return errors.Wrap(err, "failed to persist the calculated transaction clusters")
	}

	return nil
}
