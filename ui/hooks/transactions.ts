import { InfiniteData, useInfiniteQuery, UseInfiniteQueryResult, useMutation, useQueryClient } from 'react-query';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import Balance from 'models/Balance';
import Spending from 'models/Spending';
import Transaction from 'models/Transaction';
import request from 'util/request';

export type TransactionsResult =
  {
    result: Array<Transaction>;
    hasNextPage: boolean;
  }
  & UseInfiniteQueryResult<Array<Partial<Transaction>>>;

export function useTransactionsSink(): TransactionsResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useInfiniteQuery<Array<Partial<Transaction>>>(
    `/bank_accounts/${ selectedBankAccountId }/transactions`,
    {
      getNextPageParam: (_, pages) => pages.length * 25,
      keepPreviousData: true,
      enabled: !!selectedBankAccountId,
    },
  );
  return {
    ...result,
    hasNextPage: !result?.data?.pages.some(page => page.length < 25),
    // Take all the pages and build an array. Make sure we actually return an array here even if it's empty.
    result: result?.data?.pages.flatMap(x => x).map(item => new Transaction(item)) || [],
  };
}

export function useUpdateTransaction(): (_transaction: Transaction) => Promise<void> {
  const queryClient = useQueryClient();

  interface TransactionUpdateResponse {
    transaction: Partial<Transaction>;
    spending: Array<Partial<Spending>>;
    balance: Partial<Balance>;
  }

  async function updateTransaction(transaction: Transaction): Promise<TransactionUpdateResponse> {
    return request()
      .put<TransactionUpdateResponse>(
        `/bank_accounts/${ transaction.bankAccountId }/transactions/${ transaction.transactionId }`,
        transaction,
      )
      .then(result => result.data);
  }

  const { mutate } = useMutation(
    updateTransaction,
    {
      onSuccess: (response: TransactionUpdateResponse) => Promise.all([
        queryClient.setQueriesData(
          `/bank_accounts/${ response.transaction.bankAccountId }/transactions`,
          (previous: InfiniteData<Array<Transaction>>) => ({
            ...previous,
            pages: previous.pages.map(page =>
              page.map(item =>
                item.transactionId === response.transaction.transactionId ? response.transaction : item
              )
            ),
          })
        ),
        queryClient.setQueriesData(
          `/bank_accounts/${ response.transaction.bankAccountId }/spending`,
          (previous: Array<Partial<Spending>>) => previous
            .map(item => (response.spending || []).find(updated => updated.spendingId === item.spendingId) || item),
        ),
        queryClient.setQueriesData(
          `/bank_accounts/${ response.transaction.bankAccountId }/balances`,
          (previous: Partial<Balance>) => new Balance({
            ...previous,
            ...response.balance,
          }),
        ),
      ]),
    }
  );

  return async (transaction: Transaction): Promise<void> => {
    return mutate(transaction);
  };
}
