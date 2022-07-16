import { InfiniteData, useInfiniteQuery, UseInfiniteQueryResult, useMutation, useQueryClient } from 'react-query';

import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import Balance from 'models/Balance';
import Spending from 'models/Spending';
import Transaction from 'models/Transaction';
import request from 'util/request';

export type TransactionsResult =
  { result: Map<number, Transaction> }
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
    // Take all the items from all the pages and build a map with the results, keyed by the transaction ID.
    result: new Map(result?.data?.pages.flatMap(x => x).map(item => {
      const transaction = new Transaction(item);
      return [transaction.transactionId, transaction];
    })),
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
            .map(item => response.spending.find(updated => updated.spendingId === item.spendingId) || item),
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
