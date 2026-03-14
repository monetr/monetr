package jobs

import (
	"context"

	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/datasources/ofx/ofx_jobs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/similar"
	"github.com/monetr/monetr/server/storage/storage_jobs"
)

func RegisterJobs(ctx context.Context, processor queue.Processor) error {
	return myownsanity.FirstError(
		queue.Register(ctx, processor, background.CleanupLunchFlow),
		queue.Register(ctx, processor, background.DeactivateLink),
		queue.Register(ctx, processor, ofx_jobs.ProcessOFXUpload),
		queue.Register(ctx, processor, lunch_flow.SyncLunchFlow),
		queue.Register(ctx, processor, similar.CalculateTransactionClusters),
		queue.Register(ctx, processor, storage_jobs.RemoveFile),
		queue.RegisterCron(ctx, processor, background.CleanupJobsCron, "0 0 8 * * *"),
		queue.RegisterCron(ctx, processor, background.CleanupLunchFlowCron, "0 15 1 * * *"),
		queue.RegisterCron(ctx, processor, background.DeactivateLinksCron, "0 0 0 * * *"),
		queue.RegisterCron(ctx, processor, background.ProcessSpendingCron, "0 30 * * * *"),
		queue.RegisterCron(ctx, processor, lunch_flow.SyncLunchFlowCron, "0 20 */6 * * *"),
		queue.RegisterCron(ctx, processor, storage_jobs.CleanupFilesCron, "0 28 * * * *"),
	)
}
