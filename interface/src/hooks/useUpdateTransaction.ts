import { type InfiniteData, useMutation } from '@tanstack/react-query';

import Balance from '@monetr/interface/models/Balance';
import type Spending from '@monetr/interface/models/Spending';
import type Transaction from '@monetr/interface/models/Transaction';
import request from '@monetr/interface/util/request';

export interface TransactionUpdateResponse {
  transaction: Partial<Transaction>;
  spending: Array<Partial<Spending>>;
  balance: Partial<Balance>;
}

export function useUpdateTransaction(): (_transaction: Transaction) => Promise<TransactionUpdateResponse> {
  const { mutateAsync } = useMutation({
    async mutationFn(transaction: Transaction): Promise<TransactionUpdateResponse> {
      return request()
        .put<TransactionUpdateResponse>(
          `/bank_accounts/${transaction.bankAccountId}/transactions/${transaction.transactionId}`,
          transaction,
        )
        .then(result => result.data);
    },
    onSuccess: ({ transaction, spending, balance }: TransactionUpdateResponse, _input, _, { client: queryClient }) =>
      Promise.all([
        queryClient.setQueryData<InfiniteData<Array<Partial<Transaction>>>>(
          [`/bank_accounts/${transaction.bankAccountId}/transactions`],
          previous =>
            // If previous does not exist then do nothing, otherwise this will break the page.
            previous && {
              ...previous,
              // Map over all of the pages
              pages: previous.pages?.map(page =>
                // And over all the items in the page, looking for transactions with the same ID as the one we are
                // updating, if we find one then return the updated transaction instead of the one that was there.
                page.map(item => (item.transactionId === transaction.transactionId ? transaction : item)),
              ),
            },
          {},
        ),
        queryClient.setQueryData(
          [`/bank_accounts/${transaction.bankAccountId}/transactions/${transaction.transactionId}`],
          transaction,
        ),
        queryClient.setQueryData<Array<Partial<Spending>>>(
          [`/bank_accounts/${transaction.bankAccountId}/spending`],
          previous =>
            // Since there could be multiple spending objects updated here, we need to take map over all of the existing
            // spendinng objects and then check to see if that spending object is in the array of updated spending
            // objects. If it is, replace it.
            previous?.map(item => (spending || []).find(updated => updated.spendingId === item.spendingId) || item),
        ),
        // For all of the spending objects that were updated we need to make sure to update their individual items if
        // they have been requested.
        (spending || []).map(spending =>
          Promise.all([
            queryClient.setQueryData(
              [`/bank_accounts/${transaction.bankAccountId}/spending/${spending.spendingId}`],
              spending,
            ),
            queryClient.invalidateQueries({
              queryKey: [`/bank_accounts/${transaction.bankAccountId}/spending/${spending.spendingId}/transactions`],
            }),
          ]),
        ),
        queryClient.setQueryData<Partial<Balance>>(
          [`/bank_accounts/${transaction.bankAccountId}/balances`],
          previous => new Balance({ ...previous, ...balance }),
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
