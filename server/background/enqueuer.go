package background

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.4.0 -source=enqueuer.go -package=mockgen -destination=../internal/mockgen/enqueuer.go JobEnqueuer
type JobEnqueuer interface {
	EnqueueJob(ctx context.Context, queue string, data interface{}) error
}
