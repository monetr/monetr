import { useMutation, useQueryClient } from '@tanstack/react-query';
import { AxiosResponse } from 'axios';

import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export type PatchFundingScheduleRequest =
  Pick<FundingSchedule, 'fundingScheduleId' | 'bankAccountId'> &
  Partial<Pick<FundingSchedule,
    'name' |
    'description' |
    'ruleset' |
    'nextRecurrence' |
    'excludeWeekends' |
    'estimatedDeposit'
  >>

export interface PatchFundingScheduleResponse {
  fundingSchedule: FundingSchedule;
  spending: Array<Spending>;
}

export function usePatchFundingSchedule(): (_patch: PatchFundingScheduleRequest) => Promise<PatchFundingScheduleResponse> {
  const queryClient = useQueryClient();

  async function patchFundingSchedule(
    { fundingScheduleId, bankAccountId, ...patch }: PatchFundingScheduleRequest,
  ): Promise<PatchFundingScheduleResponse> {
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

  const mutation = useMutation(
    patchFundingSchedule,
    {
      onSuccess: (response: PatchFundingScheduleResponse) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => previous.map(item =>
            item.fundingScheduleId === response.fundingSchedule.fundingScheduleId ? response.fundingSchedule : item
          ),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/funding_schedules/${response.fundingSchedule.fundingScheduleId}`],
          response.fundingSchedule,
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.fundingSchedule.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => previous
            .map(item => (response.spending || []).find(updated => updated.spendingId === item.spendingId) || item),
        ),
        (response.spending || []).map(spending =>
          queryClient.setQueriesData(
            [`/bank_accounts/${response.fundingSchedule.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          )),
        queryClient.invalidateQueries([`/bank_accounts/${ response.fundingSchedule.bankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ response.fundingSchedule.bankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return mutation.mutateAsync;
}
