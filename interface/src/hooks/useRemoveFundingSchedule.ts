import { useMutation, useQueryClient } from '@tanstack/react-query';

import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import request from '@monetr/interface/util/request';

export function useRemoveFundingSchedule(): (_fundingSchedule: FundingSchedule) => Promise<FundingSchedule> {
  const queryClient = useQueryClient();

  async function removeFundingSchedule(fundingSchedule: FundingSchedule): Promise<FundingSchedule> {
    return request()
      .delete(`/bank_accounts/${fundingSchedule.bankAccountId}/funding_schedules/${fundingSchedule.fundingScheduleId}`)
      .then(() => fundingSchedule);
  }

  const mutation = useMutation({
    mutationFn: removeFundingSchedule,
    onSuccess: ({ bankAccountId, fundingScheduleId }: FundingSchedule) =>
      Promise.all([
        queryClient.setQueryData(
          [`/bank_accounts/${bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) =>
            previous.filter(item => item.fundingScheduleId !== fundingScheduleId),
        ),
        queryClient.removeQueries({
          queryKey: [`/bank_accounts/${bankAccountId}/funding_schedules/${fundingScheduleId}`],
        }),
      ]),
  });

  return mutation.mutateAsync;
}
