package mockqueue

import (
	"github.com/monetr/monetr/server/queue"
	"go.uber.org/mock/gomock"
)

// EqQueue is a helper function that returns a gomock matcher used to assert
// that a specific job was enqueued. For example:
//
//	processor := mockgen.NewMockProcessor(ctrl)
//	processor.EXPECT().
//
//	 EnqueueAt(
//		gomock.Any(),
//		mockqueue.EqQueue(similar.CalculateTransactionClusters),
//		gomock.Any(),
//		gomock.Eq(similar.CalculateTransactionClustersArguments{
//			AccountId:     bankAccount.AccountId,
//			BankAccountId: bankAccount.BankAccountId,
//		}),
//	 ).
//	 Return(nil).
//	 Times(1)
func EqQueue[T any, F func(ctx queue.Context, args T) error](
	callback F,
) gomock.Matcher {
	return gomock.Eq(queue.QueueNameFromJobFunction[T](callback))
}
