package cache

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/sirupsen/logrus"
	"time"
)

type RemovedTransactionsCache interface {
	CacheDeletedTransaction(ctx context.Context, transaction models.Transaction) bool
	LookupDeletedTransaction(ctx context.Context, transactionId string) (*models.Transaction, bool)
	Close() error
}

var (
	_ RemovedTransactionsCache = &noopRemovedTransactionsCache{}
	_ RemovedTransactionsCache = &redisRemovedTransactionsCache{}
)

type noopRemovedTransactionsCache struct{}

func (n noopRemovedTransactionsCache) CacheDeletedTransaction(ctx context.Context, transaction models.Transaction) bool {
	return false
}

func (n noopRemovedTransactionsCache) LookupDeletedTransaction(ctx context.Context, transactionId string) (*models.Transaction, bool) {
	return nil, false
}

func (n noopRemovedTransactionsCache) Close() error {
	return nil
}

type redisRemovedTransactionsCache struct {
	log    *logrus.Entry
	client Cache
}

func (r *redisRemovedTransactionsCache) CacheDeletedTransaction(ctx context.Context, transaction models.Transaction) bool {
	span := sentry.StartSpan(ctx, "CacheDeletedTransaction")
	defer span.Finish()

	return r.client.SetEzTTL(span.Context(), r.getKey(transaction), transaction, 30*time.Minute) == nil
}

func (r *redisRemovedTransactionsCache) LookupDeletedTransaction(ctx context.Context, transactionId string) (*models.Transaction, bool) {
	panic("implement me")
}

func (r *redisRemovedTransactionsCache) Close() error {
	panic("implement me")
}

func (r *redisRemovedTransactionsCache) getKey(transaction models.Transaction) string {
	return fmt.Sprintf("plaid:deleted_transactions:%d:%s", transaction.AccountId, transaction.PlaidTransactionId)
}
