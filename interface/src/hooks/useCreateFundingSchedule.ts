import { useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { ID } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

// A create only ever sends the fields the user actually filled out, the rest of the writable fields are computed by the
// server so we make the writable portion partial here. The required fields are enforced by the form validators.
export type CreateFundingScheduleRequest = Writable<FundingSchedule> & { bankAccountId: ID<BankAccount> };

export function useCreateFundingSchedule(): (_funding: CreateFundingScheduleRequest) => Promise<FundingSchedule> {
  const createFundingSchedule = useCallback(
    async ({ bankAccountId, ...fundingSchedule }: CreateFundingScheduleRequest): Promise<FundingSchedule> => {
      return await request<WithJsonValues<FundingSchedule>>({
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
          ['GET', `/api/bank_accounts/${data.bankAccountId}/funding_schedules`],
          // If the list hasn't been loaded yet then leave it alone, otherwise we'd cache a list with only this one in it
          (previous: Array<Partial<FundingSchedule>> | undefined) => previous?.concat(data),
        ),
        ctx.client.setQueryData(
          ['GET', `/api/bank_accounts/${data.bankAccountId}/funding_schedules/${data.fundingScheduleId}`],
          data,
        ),
      ]),
  });

  return mutateAsync;
}
