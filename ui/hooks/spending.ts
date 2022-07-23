import { useMutation, useQuery, useQueryClient, UseQueryResult } from 'react-query';
import shallow from 'zustand/shallow';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import useStore from 'hooks/store';
import Spending, { SpendingType } from 'models/Spending';
import request from 'shared/util/request';

export type SpendingResult =
  { result: Map<number, Spending> }
  & UseQueryResult<Array<Partial<Spending>>>;

export function useSpendingSink(): SpendingResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useQuery<Array<Partial<Spending>>>(
    `/bank_accounts/${ selectedBankAccountId }/spending`,
    {
      enabled: !!selectedBankAccountId,
    },
  );
  return {
    ...result,
    result: new Map(result?.data?.map(item => {
      const spending = new Spending(item);
      return [spending.spendingId, spending];
    })),
  };
}

export function useSpending(spendingId: number): Spending | null {
  const { result } = useSpendingSink();
  return result.get(spendingId) || null;
}

export function useSpendingFiltered(kind: SpendingType): SpendingResult {
  const base = useSpendingSink();
  return {
    ...base,
    result: new Map(
      Array.from(base.result.values())
        .filter(item => item.spendingType === kind)
        .map(item => [item.spendingId, item]),
    ),
  };
}

export function useSelectedExpense(): Spending | null {
  const selectedExpenseId: number | null = useStore(state => state.selectedBankAccountId, shallow);
  const { isLoading, result } = useSpendingFiltered(SpendingType.Expense);
  if (isLoading) return null;

  return result.get(selectedExpenseId) || null;
}

export function useRemoveSpending(): (_spendingId: number) => Promise<void> {
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  async function removeSpending(spendingId: number): Promise<number> {
    return request()
      .delete(`/bank_accounts/${ selectedBankAccountId }/spending/${ spendingId }`)
      .then(() => spendingId);
  }

  const { mutate } = useMutation(
    removeSpending,
    {
      onSuccess: (removedSpendingId: number) => Promise.all([
        queryClient.setQueriesData(
          `/bank_accounts/${ selectedBankAccountId }/spending`,
          (previous: Array<Partial<Spending>>) => previous.filter(item => item.spendingId !== removedSpendingId),
        ),
        queryClient.invalidateQueries(`/bank_accounts/${ selectedBankAccountId }/balances`),
      ]),
    },
  );

  return async (spendingId: number): Promise<void> => {
    return mutate(spendingId);
  };
}

export function useUpdateSpending(): (_spending: Spending) => Promise<void> {
  const queryClient = useQueryClient();

  async function updateSpending(spending: Spending): Promise<Spending> {
    return request()
      .put<Partial<Spending>>(`/bank_accounts/${ spending.bankAccountId }/spending/${ spending.spendingId }`, spending)
      .then(result => new Spending(result?.data));
  }

  const { mutate } = useMutation(
    updateSpending,
    {
      onSuccess: (updatedSpending: Spending) => Promise.all([
        queryClient.setQueriesData(
          `/bank_accounts/${ updatedSpending.bankAccountId }/spending`,
          (previous: Array<Partial<Spending>>) =>
            previous.map(item => item.spendingId === updatedSpending.spendingId ? updatedSpending : item),
        ),
        queryClient.invalidateQueries(`/bank_accounts/${ updatedSpending.bankAccountId }/balances`),
      ]),
    },
  );

  return async (spending: Spending): Promise<void> => {
    return mutate(spending);
  };
}

export function useCreateSpending(): (_spending: Spending) => Promise<void> {
  const queryClient = useQueryClient();

  async function createSpending(spending: Spending): Promise<Spending> {
    return request()
      .post<Partial<Spending>>(`/bank_accounts/${ spending.bankAccountId }/spending`, spending)
      .then(result => new Spending(result?.data));
  }

  const { mutate } = useMutation(
    createSpending,
    {
      onSuccess: (createdSpending: Spending) => Promise.all([
        queryClient.setQueriesData(
          `/bank_accounts/${ createdSpending.bankAccountId }/spending`,
          (previous: Array<Partial<Spending>>) => previous.concat(createdSpending),
        ),
        queryClient.invalidateQueries(`/bank_accounts/${ createdSpending.bankAccountId }/balances`),
      ]),
    },
  );

  return async (spending: Spending): Promise<void> => {
    return mutate(spending);
  };
}
