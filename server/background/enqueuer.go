package background

import "context"

//go:generate mockgen -source=enqueuer.go -package=mockgen -destination=../internal/mockgen/enqueuer.go JobEnqueuer
type JobEnqueuer interface {
	EnqueueJob(ctx context.Context, queue string, data interface{}) error
}
