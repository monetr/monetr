import { useMutation, useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import Balance from '@monetr/interface/models/Balance';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export type SpendingResult =
  { result: Array<Spending> }
  & UseQueryResult<Array<Partial<Spending>>>;

export function useSpendings(): UseQueryResult<Array<Spending>> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<Spending>>, unknown, Array<Spending>>(
    [`/bank_accounts/${ selectedBankAccountId }/spending`],
    {
      enabled: !!selectedBankAccountId,
      initialData: [],
      initialDataUpdatedAt: 0,
      select: data => (data || []).map(item => new Spending(item)),
    },
  );
}

export function useSpending(spendingId: string | null): UseQueryResult<Spending> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Spending>, unknown, Spending>(
    [`/bank_accounts/${ selectedBankAccountId }/spending/${spendingId}`],
    {
      enabled: !!selectedBankAccountId,
      select: data => new Spending(data),
    },
  );
}

/**
 * useSpending retrieves a single spending item that would have been returned from the index endpoint for the currently
 * selected bank account.
 */
export function useSpendingOld(spendingId?: string): Spending | null {
  const { data } = useSpendings();
  if (!spendingId || !data) {
    return null;
  }

  return data?.find(item => item.spendingId === spendingId) || null;
}

export function useSpendingFiltered(kind: SpendingType): SpendingResult {
  const base = useSpendings();
  return {
    ...base,
    result: base.data?.filter(item => item.spendingType === kind) || [],
  };
}

export function useRemoveSpending(): (_spendingId: string) => Promise<void> {
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  async function removeSpending(spendingId: string): Promise<string> {
    return request()
      .delete(`/bank_accounts/${ selectedBankAccountId }/spending/${ spendingId }`)
      .then(() => spendingId);
  }

  const { mutate } = useMutation(
    removeSpending,
    {
      onSuccess: (removedSpendingId: string) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${ selectedBankAccountId }/spending`],
          (previous: Array<Partial<Spending>>) => previous.filter(item => item.spendingId !== removedSpendingId),
        ),
        queryClient.removeQueries([`/bank_accounts/${ selectedBankAccountId }/spending/${ removedSpendingId }`]),
        queryClient.invalidateQueries([`/bank_accounts/${ selectedBankAccountId }/balances`]),
        queryClient.invalidateQueries([`/bank_accounts/${ selectedBankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ selectedBankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return async (spendingId: string): Promise<void> => {
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
          [`/bank_accounts/${ updatedSpending.bankAccountId }/spending`],
          (previous: Array<Partial<Spending>>) =>
            previous.map(item => item.spendingId === updatedSpending.spendingId ? updatedSpending : item),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${ updatedSpending.bankAccountId}/spending/${ updatedSpending.spendingId}`],
          updatedSpending,
        ),
        // TODO Under what circumstances do we need to invalidate balances for a spending update?
        queryClient.invalidateQueries([`/bank_accounts/${ updatedSpending.bankAccountId }/balances`]),
        queryClient.invalidateQueries([`/bank_accounts/${ updatedSpending.bankAccountId }/forecast`]),
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
          [`/bank_accounts/${ createdSpending.bankAccountId }/spending`],
          (previous: Array<Partial<Spending>>) => (previous || []).concat(createdSpending),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${ createdSpending.bankAccountId}/spending/${ createdSpending.spendingId}`],
          createdSpending,
        ),
        queryClient.invalidateQueries([`/bank_accounts/${ createdSpending.bankAccountId }/balances`]),
        queryClient.invalidateQueries([`/bank_accounts/${ createdSpending.bankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ createdSpending.bankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return async (spending: Spending): Promise<Spending> => {
    return mutation.mutateAsync(spending);
  };
}

export type TransferParameters = {
  fromSpendingId: string | null,
  toSpendingId: string | null,
  amount: number,
}

export function useTransfer(): (transferParameters: TransferParameters) => Promise<any> {
  const queryClient = useQueryClient();

  interface BalanceTransferResponse {
    balance: Partial<Balance>;
    spending: Array<Partial<Spending>>;
  }

  interface BalanceTransferRequest {
    fromSpendingId: string | null;
    toSpendingId: string | null;
    amount: number;
  }

  const selectedBankAccountId = useSelectedBankAccountId();

  async function transfer(transferRequest: BalanceTransferRequest): Promise<BalanceTransferResponse> {
    return request()
      .post<BalanceTransferResponse>(`/bank_accounts/${ selectedBankAccountId }/spending/transfer`, transferRequest)
      .then(result => result.data);
  }

  const { mutateAsync } = useMutation(
    transfer,
    {
      onSuccess: (result: BalanceTransferResponse) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${ selectedBankAccountId }/spending`],
          (previous: Array<Partial<Spending>>) => previous
            .map(item => result.spending.find(updated => updated.spendingId === item.spendingId) || item),
        ),
        result.spending.map(updatedSpending => queryClient.setQueriesData(
          [`/bank_accounts/${ selectedBankAccountId }/spending/${updatedSpending.spendingId}`],
          () => updatedSpending,
        )),
        queryClient.setQueriesData(
          [`/bank_accounts/${ selectedBankAccountId }/balances`],
          (previous: Partial<Balance>) => new Balance({
            ...previous,
            ...result.balance,
          }),
        ),
        queryClient.invalidateQueries([`/bank_accounts/${ selectedBankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ selectedBankAccountId }/forecast/next_funding`]),
      ]),
    },
  );

  return mutateAsync;
}
