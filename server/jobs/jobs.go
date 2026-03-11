package jobs

import (
	"context"

	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/queue"
)

func RegisterJobs(ctx context.Context, processor queue.Processor) error {
	return myownsanity.FirstError(
		queue.Register(ctx, processor, background.CalculateTransactionClusters),
		queue.Register(ctx, processor, background.CleanupLunchFlow),
		queue.Register(ctx, processor, background.DeactivateLink),
		queue.Register(ctx, processor, background.ProcessOFXUpload),
		queue.Register(ctx, processor, background.RemoveFile),
		queue.RegisterCron(ctx, processor, "0 0 0 * * *", background.DeactivateLinksCron),
		queue.RegisterCron(ctx, processor, "0 0 8 * * *", background.CleanupJobsCron),
		queue.RegisterCron(ctx, processor, "0 15 1 * * *", background.CleanupLunchFlowCron),
		queue.RegisterCron(ctx, processor, "0 28 * * * *", background.CleanupFilesCron),
		queue.RegisterCron(ctx, processor, "0 30 * * * *", background.ProcessSpendingCron),
	)
}
