package jobs

import (
	"context"

	"github.com/monetr/monetr/server/billing/billing_jobs"
	"github.com/monetr/monetr/server/datasources/lunch_flow/lunch_flow_jobs"
	"github.com/monetr/monetr/server/datasources/ofx/ofx_jobs"
	"github.com/monetr/monetr/server/datasources/plaid/plaid_jobs"
	"github.com/monetr/monetr/server/funding/funding_jobs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/similar"
	"github.com/monetr/monetr/server/spending/spending_jobs"
	"github.com/monetr/monetr/server/storage/storage_jobs"
)

func RegisterJobs(ctx context.Context, processor queue.Processor) error {
	return myownsanity.FirstError(
		queue.Register(ctx, processor, billing_jobs.NotificationTrialExpiry),
		queue.Register(ctx, processor, billing_jobs.ReconcileSubscription),
		queue.Register(ctx, processor, funding_jobs.ProcessFundingSchedule),
		queue.Register(ctx, processor, lunch_flow_jobs.CleanupLunchFlow),
		queue.Register(ctx, processor, lunch_flow_jobs.SyncLunchFlow),
		queue.Register(ctx, processor, ofx_jobs.ProcessOFXUpload),
		queue.Register(ctx, processor, plaid_jobs.DeactivatePlaidLink),
		queue.Register(ctx, processor, plaid_jobs.SyncPlaid),
		queue.Register(ctx, processor, similar.CalculateTransactionClusters),
		queue.Register(ctx, processor, spending_jobs.ProcessSpending),
		queue.Register(ctx, processor, storage_jobs.RemoveFile),
		queue.RegisterCron(ctx, processor, CleanupJobsCron, "0 0 8 * * *"),
		queue.RegisterCron(ctx, processor, billing_jobs.NotificationTrialExpiryCron, "0 30 */6 * * *"),
		queue.RegisterCron(ctx, processor, billing_jobs.ReconcileSubscriptionCron, "0 15 */12 * * *"),
		queue.RegisterCron(ctx, processor, funding_jobs.ProcessFundingSchedulesCron, "0 0 * * * *"),
		queue.RegisterCron(ctx, processor, lunch_flow_jobs.CleanupLunchFlowCron, "0 15 1 * * *"),
		queue.RegisterCron(ctx, processor, lunch_flow_jobs.SyncLunchFlowCron, "0 20 */6 * * *"),
		queue.RegisterCron(ctx, processor, plaid_jobs.DeactivatePlaidLinkCron, "0 0 0 * * *"),
		queue.RegisterCron(ctx, processor, plaid_jobs.SyncPlaidCron, "0 0 */12 * * *"),
		queue.RegisterCron(ctx, processor, spending_jobs.ProcessSpendingCron, "0 30 * * * *"),
		queue.RegisterCron(ctx, processor, storage_jobs.CleanupFilesCron, "0 28 * * * *"),
	)
}
