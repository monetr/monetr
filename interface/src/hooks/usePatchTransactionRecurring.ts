import { useMutation, useQueryClient } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { WithJsonValues } from '@monetr/interface/util/json';
import request from '@monetr/interface/util/request';

// Only the links on a recurring transaction can be changed, everything else about it is calculated by the server.
export interface PatchTransactionRecurringRequest {
  transactionRecurringId: ID<TransactionRecurring>;
  bankAccountId: ID<BankAccount>;
  spendingId?: ID<Spending> | null;
  fundingScheduleId?: ID<FundingSchedule> | null;
}

export function usePatchTransactionRecurring(): (_: PatchTransactionRecurringRequest) => Promise<TransactionRecurring> {
  const queryClient = useQueryClient();

  async function patchTransactionRecurring({
    transactionRecurringId,
    bankAccountId,
    ...patch
  }: PatchTransactionRecurringRequest): Promise<TransactionRecurring> {
    return await request<WithJsonValues<TransactionRecurring>>({
      method: 'PATCH',
      url: `/api/bank_accounts/${bankAccountId}/recurring/${transactionRecurringId}`,
      // undefined keys get dropped when this is stringified, a null still gets sent so the link can be cleared.
      data: patch,
    }).then(result => new TransactionRecurring(result.data));
  }

  const mutation = useMutation({
    mutationFn: patchTransactionRecurring,
    // the recurring transaction embeds whatever its linked to, so refresh it to pick up the new link
    onSuccess: (data: TransactionRecurring) =>
      queryClient.invalidateQueries({
        queryKey: [`/api/bank_accounts/${data.bankAccountId}/recurring/${data.transactionRecurringId}`],
      }),
  });

  return mutation.mutateAsync;
}
