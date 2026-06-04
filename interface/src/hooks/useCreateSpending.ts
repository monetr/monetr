import { useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';

import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import Spending, { type SpendingType } from '@monetr/interface/models/Spending';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type CreateSpendingRequest = Writable<Spending> & {
  bankAccountId: ID<BankAccount>;
  spendingType: SpendingType;
};

export function useCreateSpending(): (_spending: CreateSpendingRequest) => Promise<Spending> {
  const createSpending = useCallback(
    async ({ bankAccountId, ...spending }: CreateSpendingRequest): Promise<Spending> => {
      return request<WithJsonValues<Spending>>({
        method: 'POST',
        url: `/api/bank_accounts/${bankAccountId}/spending`,
        data: spending,
      }).then(result => new Spending(result.data));
    },
    [],
  );

  const { mutateAsync } = useMutation({
    mutationFn: createSpending,
    onSuccess: (data: Spending, _var, _result, ctx) =>
      Promise.all([
        ctx.client.setQueryData(
          [`/api/bank_accounts/${data.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => (previous || []).concat(data),
        ),
        ctx.client.setQueryData([`/api/bank_accounts/${data.bankAccountId}/spending/${data.spendingId}`], data),
        ctx.client.invalidateQueries({ queryKey: [`/api/bank_accounts/${data.bankAccountId}/balances`] }),
        ctx.client.invalidateQueries({ queryKey: [`/api/bank_accounts/${data.bankAccountId}/forecast`] }),
        ctx.client.invalidateQueries({
          queryKey: [`/api/bank_accounts/${data.bankAccountId}/forecast/next_funding`],
        }),
      ]),
  });

  return mutateAsync;
}
