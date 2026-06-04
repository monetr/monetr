import { useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { ID } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type CreateFundingScheduleRequest = Writable<FundingSchedule> & { bankAccountId: ID<BankAccount> };

export function useCreateFundingSchedule(): (_funding: CreateFundingScheduleRequest) => Promise<FundingSchedule> {
  const createFundingSchedule = useCallback(
    async ({ bankAccountId, ...fundingSchedule }: CreateFundingScheduleRequest): Promise<FundingSchedule> => {
      return request<WithJsonValues<FundingSchedule>>({
        method: 'POST',
        url: `/api/bank_accounts/${bankAccountId}/funding_schedules`,
        data: fundingSchedule,
      }).then(result => new FundingSchedule(result.data));
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
