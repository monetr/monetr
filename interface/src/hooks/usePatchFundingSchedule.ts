import { useMutation, useQueryClient } from '@tanstack/react-query';
import { AxiosResponse } from 'axios';

import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export type PatchFundingScheduleRequest = Pick<FundingSchedule, 'fundingScheduleId' | 'bankAccountId'> &
  Partial<
    Pick<
      FundingSchedule,
      'name' | 'description' | 'ruleset' | 'nextRecurrence' | 'excludeWeekends' | 'estimatedDeposit'
    >
  >;

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
    return request()
      .patch<FundingSchedule, AxiosResponse<PatchFundingScheduleResponse>>(
        `/bank_accounts/${bankAccountId}/funding_schedules/${fundingScheduleId}`,
        patch,
      )
      .then(result => ({
        fundingSchedule: new FundingSchedule(result.data.fundingSchedule),
        spending: result.data.spending.map(spending => new Spending(spending)),
      }));
  }

  const mutation = useMutation({
    mutationFn: patchFundingSchedule,
    onSuccess: ({ fundingSchedule, spending }: PatchFundingScheduleResponse) =>
      Promise.all([
        queryClient.setQueryData(
          [`/bank_accounts/${fundingSchedule.bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) =>
            (previous ?? []).map(item =>
              item.fundingScheduleId === fundingSchedule.fundingScheduleId ? fundingSchedule : item,
            ),
        ),
        queryClient.setQueryData(
          [`/bank_accounts/${fundingSchedule.bankAccountId}/funding_schedules/${fundingSchedule.fundingScheduleId}`],
          fundingSchedule,
        ),
        queryClient.setQueryData(
          [`/bank_accounts/${fundingSchedule.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            (previous ?? []).map(
              item => (spending || []).find(updated => updated.spendingId === item.spendingId) || item,
            ),
        ),
        (spending || []).map(spending =>
          queryClient.setQueryData(
            [`/bank_accounts/${fundingSchedule.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          ),
        ),
        queryClient.invalidateQueries({
          queryKey: [`/bank_accounts/${fundingSchedule.bankAccountId}/forecast`],
        }),
        queryClient.invalidateQueries({
          queryKey: [`/bank_accounts/${fundingSchedule.bankAccountId}/forecast/next_funding`],
        }),
      ]),
  });

  return mutation.mutateAsync;
}
