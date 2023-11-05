import {
  InfiniteData,
  useInfiniteQuery,
  UseInfiniteQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
  UseQueryResult,
} from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import Balance from '@monetr/interface/models/Balance';
import Spending from '@monetr/interface/models/Spending';
import Transaction from '@monetr/interface/models/Transaction';
import request from '@monetr/interface/util/request';

export type TransactionsResult = {
  result: Array<Transaction>;
  hasNextPage: boolean;
} & UseInfiniteQueryResult<Array<Partial<Transaction>>>;

export function useTransactions(): TransactionsResult {
  const selectedBankAccountId = useSelectedBankAccountId();
  const result = useInfiniteQuery<Array<Partial<Transaction>>>(
    [`/bank_accounts/${ selectedBankAccountId }/transactions`],
    {
      getNextPageParam: (_, pages) => pages.length * 25,
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

export function useTransaction(transactionId: number | null): UseQueryResult<Transaction> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Transaction>, unknown, Transaction>(
    [`/bank_accounts/${ selectedBankAccountId }/transactions/${ transactionId }`],
    {
      enabled: !!selectedBankAccountId && !!transactionId,
      select: data => new Transaction(data),
    },
  );
}

export interface TransactionUpdateResponse {
  transaction: Partial<Transaction>;
  spending: Array<Partial<Spending>>;
  balance: Partial<Balance>;
}

export function useUpdateTransaction(): (_transaction: Transaction) => Promise<TransactionUpdateResponse> {
  const queryClient = useQueryClient();

  async function updateTransaction(transaction: Transaction): Promise<TransactionUpdateResponse> {
    return request()
      .put<TransactionUpdateResponse>(
        `/bank_accounts/${ transaction.bankAccountId }/transactions/${ transaction.transactionId }`,
        transaction,
      )
      .then(result => result.data);
  }

  const { mutateAsync } = useMutation(
    updateTransaction,
    {
      onSuccess: (response: TransactionUpdateResponse) => Promise.all([
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/transactions`],
          (previous: InfiniteData<Array<Transaction>>) => ({
            ...previous,
            pages: previous.pages.map(page =>
              page.map(item =>
                item.transactionId === response.transaction.transactionId ? response.transaction : item
              )
            ),
          }),
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/transactions/${response.transaction.transactionId}`],
          response.transaction,
        ),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) => previous
            .map(item => (response.spending || []).find(updated => updated.spendingId === item.spendingId) || item),
        ),
        (response.spending || []).map(spending =>
          queryClient.setQueriesData(
            [`/bank_accounts/${response.transaction.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          )),
        queryClient.setQueriesData(
          [`/bank_accounts/${response.transaction.bankAccountId}/balances`],
          (previous: Partial<Balance>) => new Balance({
            ...previous,
            ...response.balance,
          }),
        ),
        queryClient.invalidateQueries([`/bank_accounts/${ response.transaction.bankAccountId }/forecast`]),
        queryClient.invalidateQueries([`/bank_accounts/${ response.transaction.bankAccountId }/forecast/next_funding`]),
      ]),
    }
  );

  return mutateAsync;
}
