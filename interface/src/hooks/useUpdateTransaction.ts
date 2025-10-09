import { InfiniteData, useMutation, useQueryClient } from '@tanstack/react-query';

import Balance from '@monetr/interface/models/Balance';
import Spending from '@monetr/interface/models/Spending';
import Transaction from '@monetr/interface/models/Transaction';
import request from '@monetr/interface/util/request';

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

  const { mutateAsync } = useMutation({
    mutationFn: updateTransaction,
    onSuccess: ({ transaction, spending, balance }: TransactionUpdateResponse) => Promise.all([
      queryClient.setQueryData(
        [`/bank_accounts/${transaction.bankAccountId}/transactions`],
        (previous: InfiniteData<Array<Transaction>>) => ({
          ...previous,
          pages: previous.pages.map(page =>
            page.map(item =>
              item.transactionId === transaction.transactionId ? transaction : item
            )
          ),
        }),
      ),
      queryClient.setQueryData(
        [`/bank_accounts/${transaction.bankAccountId}/transactions/${transaction.transactionId}`],
        transaction,
      ),
      queryClient.setQueryData(
        [`/bank_accounts/${transaction.bankAccountId}/spending`],
        (previous: Array<Partial<Spending>>) => previous
          .map(item => (spending || []).find(updated => updated.spendingId === item.spendingId) || item),
      ),
      (spending || []).map(spending =>
        queryClient.setQueryData(
          [`/bank_accounts/${transaction.bankAccountId}/spending/${spending.spendingId}`],
          spending,
        )),
      queryClient.setQueryData(
        [`/bank_accounts/${transaction.bankAccountId}/balances`],
        (previous: Partial<Balance>) => new Balance({
          ...previous,
          ...balance,
        }),
      ),
      queryClient.invalidateQueries({
        queryKey: [`/bank_accounts/${transaction.bankAccountId}/forecast`],
      }),
      queryClient.invalidateQueries({
        queryKey: [`/bank_accounts/${transaction.bankAccountId}/forecast/next_funding`],
      }),
    ]),
  });

  return mutateAsync;
}
