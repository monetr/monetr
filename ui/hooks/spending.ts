import { useMutation, useQuery, useQueryClient, UseQueryResult } from 'react-query';
import shallow from 'zustand/shallow';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import useStore from 'hooks/store';
import Balance from 'models/Balance';
import Spending, { SpendingType } from 'models/Spending';
import request from 'util/request';

export type SpendingResult =
  { result: Array<Spending> }
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
    result: (result?.data || []).map(item => new Spending(item)),
  };
}

/**
 * useSpending retrieves a single spending item that would have been returned from the index endpoint for the currently
 * selected bank account.
 *
 * @deprecated A better implementation of this will be available soon.
 */
export function useSpending(spendingId?: number): Spending | null {
  const { result } = useSpendingSink();
  if (!spendingId) {
    return null;
  }

  return result.find(item => item.spendingId === spendingId) || null;
}

export function useSpendingFiltered(kind: SpendingType): SpendingResult {
  const base = useSpendingSink();
  return {
    ...base,
    result: base.result.filter(item => item.spendingType === kind),
  };
}

export function useSelectedExpense(): Spending | null {
  const selectedExpenseId: number | null = useStore(state => state.selectedExpenseId, shallow);
  const { isLoading, result } = useSpendingFiltered(SpendingType.Expense);
  if (isLoading) return null;

  return result.find(item => item.spendingId === selectedExpenseId) || null;
}

export function useSelectedGoal(): Spending | null {
  const selectedGoalId: number | null = useStore(state => state.selectedGoalId, shallow);
  const { isLoading, result } = useSpendingFiltered(SpendingType.Goal);
  if (isLoading) return null;

  return result.find(item => item.spendingId === selectedGoalId) || null;
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
        queryClient.invalidateQueries([`/bank_accounts/${ selectedBankAccountId }/forecast/next_funding`]),
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
        queryClient.invalidateQueries([`/bank_accounts/${ updatedSpending.bankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return async (spending: Spending): Promise<void> => {
    return mutate(spending);
  };
}

export function useCreateSpending(): (_spending: Spending) => Promise<Spending> {
  const queryClient = useQueryClient();

  async function createSpending(spending: Spending): Promise<Spending> {
    return request()
      .post<Partial<Spending>>(`/bank_accounts/${ spending.bankAccountId }/spending`, spending)
      .then(result => new Spending(result?.data));
  }

  const mutation = useMutation(
    createSpending,
    {
      onSuccess: (createdSpending: Spending) => Promise.all([
        queryClient.setQueriesData(
          `/bank_accounts/${ createdSpending.bankAccountId }/spending`,
          (previous: Array<Partial<Spending>>) => (previous || []).concat(createdSpending),
        ),
        queryClient.invalidateQueries(`/bank_accounts/${ createdSpending.bankAccountId }/balances`),
        queryClient.invalidateQueries([`/bank_accounts/${ createdSpending.bankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return async (spending: Spending): Promise<Spending> => {
    return mutation.mutateAsync(spending);
  };
}

export function useTransfer(): (
  _fromSpendingId: number | null,
  _toSpendingId: number | null,
  amount: number,
) => Promise<void> {
  const queryClient = useQueryClient();

  interface BalanceTransferResponse {
    balance: Partial<Balance>;
    spending: Array<Partial<Spending>>;
  }

  interface BalanceTransferRequest {
    fromSpendingId: number | null;
    toSpendingId: number | null;
    amount: number;
  }

  const selectedBankAccountId = useSelectedBankAccountId();

  async function transfer(transferRequest: BalanceTransferRequest): Promise<BalanceTransferResponse> {
    return request()
      .post<BalanceTransferResponse>(`/bank_accounts/${ selectedBankAccountId }/spending/transfer`, transferRequest)
      .then(result => result.data);
  }

  const { mutate } = useMutation(
    transfer,
    {
      onSuccess: (result: BalanceTransferResponse) => Promise.all([
        queryClient.setQueriesData(
          `/bank_accounts/${ selectedBankAccountId }/spending`,
          (previous: Array<Partial<Spending>>) => previous
            .map(item => result.spending.find(updated => updated.spendingId === item.spendingId) || item),
        ),
        queryClient.setQueriesData(
          `/bank_accounts/${ selectedBankAccountId }/balances`,
          (previous: Partial<Balance>) => new Balance({
            ...previous,
            ...result.balance,
          }),
        ),
        queryClient.invalidateQueries([`/bank_accounts/${ selectedBankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return async (
    fromSpendingId: number | null,
    toSpendingId: number | null,
    amount: number,
  ): Promise<void> => {
    return mutate({
      fromSpendingId,
      toSpendingId,
      amount,
    });
  };
}
