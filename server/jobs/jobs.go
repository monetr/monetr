package jobs

import (
	"context"

	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/similar"
)

func RegisterJobs(ctx context.Context, processor queue.Processor) error {
	return myownsanity.FirstError(
		queue.Register(ctx, processor, background.CleanupLunchFlow),
		queue.Register(ctx, processor, background.DeactivateLink),
		queue.Register(ctx, processor, background.ProcessOFXUpload),
		queue.Register(ctx, processor, background.RemoveFile),
		queue.Register(ctx, processor, lunch_flow.SyncLunchFlow),
		queue.Register(ctx, processor, similar.CalculateTransactionClusters),
		queue.RegisterCron(ctx, processor, background.CleanupFilesCron, "0 28 * * * *"),
		queue.RegisterCron(ctx, processor, background.CleanupJobsCron, "0 0 8 * * *"),
		queue.RegisterCron(ctx, processor, background.CleanupLunchFlowCron, "0 15 1 * * *"),
		queue.RegisterCron(ctx, processor, background.DeactivateLinksCron, "0 0 0 * * *"),
		queue.RegisterCron(ctx, processor, background.ProcessSpendingCron, "0 30 * * * *"),
		queue.RegisterCron(ctx, processor, lunch_flow.SyncLunchFlowCron, "0 20 */6 * * *"),
	)
}
