import { useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';

import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type CreateFundingScheduleRequest = Writable<FundingSchedule> & { bankAccountId: string };

export function useCreateFundingSchedule(): (_funding: CreateFundingScheduleRequest) => Promise<FundingSchedule> {
  const createFundingSchedule = useCallback(
    async ({ bankAccountId, ...fundingSchedule }: CreateFundingScheduleRequest): Promise<FundingSchedule> => {
      const result = await request<WithJsonValues<FundingSchedule>>({
        method: 'POST',
        url: `/api/bank_accounts/${bankAccountId}/funding_schedules`,
        data: fundingSchedule,
      });
      return new FundingSchedule(result.data);
    },
    [],
  );

  const { mutateAsync } = useMutation({
    mutationFn: createFundingSchedule,
    onSuccess: (data: FundingSchedule, _var, _result, ctx) =>
      Promise.all([
        ctx.client.setQueryData(
          [`/api/bank_accounts/${data.bankAccountId}/funding_schedules`],
          (previous: Array<Partial<FundingSchedule>>) => (previous ?? []).concat(data),
        ),
        ctx.client.setQueryData(
          [`/api/bank_accounts/${data.bankAccountId}/funding_schedules/${data.fundingScheduleId}`],
          data,
        ),
      ]),
  });

  return mutateAsync;
}
