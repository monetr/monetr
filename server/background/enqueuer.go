package background

import (
	"context"

	"github.com/go-pg/pg/v10"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=enqueuer.go -package=mockgen -destination=../internal/mockgen/enqueuer.go JobEnqueuer
type JobEnqueuer interface {
	// Deprecated: Use EnqueueJobTxn instead.
	EnqueueJob(ctx context.Context, queue string, data interface{}) error
	// EnqueueJobTxn will create and enqueue a job inside the current provided
	// database session. This allows you to enqueue a job specifically within the
	// current transaction. So if the transaction fails, the job will not have
	// been created.
	EnqueueJobTxn(ctx context.Context, txn pg.DBI, queue string, data interface{}) error
}
