import { useMutation, useQueryClient } from '@tanstack/react-query';

import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import request from '@monetr/interface/util/request';

export type CreateFundingScheduleRequest = Pick<FundingSchedule,
  'bankAccountId' |
  'name' |
  'description' |
  'ruleset' |
  'nextRecurrence' |
  'excludeWeekends' |
  'estimatedDeposit'
>;

export function useCreateFundingSchedule(): (_funding: CreateFundingScheduleRequest) => Promise<FundingSchedule> {
  const queryClient = useQueryClient();

  async function createFundingSchedule({
    bankAccountId,
    ...fundingSchedule
  }: CreateFundingScheduleRequest): Promise<FundingSchedule> {
    return request()
      .post<Partial<FundingSchedule>>(`/bank_accounts/${bankAccountId}/funding_schedules`, fundingSchedule)
      .then(result => new FundingSchedule(result?.data));
  }

  const mutate = useMutation({
    mutationFn: createFundingSchedule,
    onSuccess: (newFunding: FundingSchedule) => Promise.all([
      queryClient.setQueryData(
        [`/bank_accounts/${newFunding.bankAccountId}/funding_schedules`],
        (previous: Array<Partial<FundingSchedule>>) => previous.concat(newFunding),
      ),
      queryClient.setQueryData(
        [`/bank_accounts/${newFunding.bankAccountId}/funding_schedules/${newFunding.fundingScheduleId}`],
        newFunding,
      ),
    ]),
  });

  return mutate.mutateAsync;
}
