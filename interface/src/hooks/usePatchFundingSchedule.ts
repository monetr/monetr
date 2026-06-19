import { useMutation, useQueryClient } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { ID } from '@monetr/interface/models/ID';
import Spending from '@monetr/interface/models/Spending';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type PatchFundingScheduleRequest = Partial<Writable<FundingSchedule>> & {
  fundingScheduleId: ID<FundingSchedule>;
  bankAccountId: ID<BankAccount>;
};

export interface PatchFundingScheduleResponse {
  fundingSchedule: FundingSchedule;
  spending: Array<Spending>;
}

export function usePatchFundingSchedule(): (_: PatchFundingScheduleRequest) => Promise<PatchFundingScheduleResponse> {
  const queryClient = useQueryClient();

  async function patchFundingSchedule({
    fundingScheduleId,
    bankAccountId,
    ...patch
  }: PatchFundingScheduleRequest): Promise<PatchFundingScheduleResponse> {
    return await request<PatchFundingScheduleResponse>({
      method: 'PATCH',
      url: `/api/bank_accounts/${bankAccountId}/funding_schedules/${fundingScheduleId}`,
      // We only want to send the fields the caller actually specified. We dont have to filter anything ourselves though,
      // the request wrapper runs this through JSON.stringify which already drops any keys that are undefined. A null
      // still gets sent because thats the caller saying they want to clear the field out.
      data: patch,
    }).then(result => ({
      fundingSchedule: new FundingSchedule(result.data.fundingSchedule),
      spending: result.data.spending.map(spending => new Spending(spending)),
    }));
  }

  const mutation = useMutation({
    mutationFn: patchFundingSchedule,
    onSuccess: ({ fundingSchedule, spending }: PatchFundingScheduleResponse) =>
      Promise.all([
        queryClient.setQueryData(
          [`/api/bank_accounts/${fundingSchedule.bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) =>
            (previous ?? []).map(item =>
              item.fundingScheduleId === fundingSchedule.fundingScheduleId ? fundingSchedule : item,
            ),
        ),
        queryClient.setQueryData(
          [
            `/api/bank_accounts/${fundingSchedule.bankAccountId}/funding_schedules/${fundingSchedule.fundingScheduleId}`,
          ],
          fundingSchedule,
        ),
        queryClient.setQueryData(
          [`/api/bank_accounts/${fundingSchedule.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            (previous ?? []).map(
              item => (spending || []).find(updated => updated.spendingId === item.spendingId) || item,
            ),
        ),
        (spending || []).map(spending =>
          queryClient.setQueryData(
            [`/api/bank_accounts/${fundingSchedule.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          ),
        ),
        queryClient.invalidateQueries({
          queryKey: [`/api/bank_accounts/${fundingSchedule.bankAccountId}/forecast`],
        }),
        queryClient.invalidateQueries({
          queryKey: [`/api/bank_accounts/${fundingSchedule.bankAccountId}/forecast/next_funding`],
        }),
      ]),
  });

  return mutation.mutateAsync;
}
