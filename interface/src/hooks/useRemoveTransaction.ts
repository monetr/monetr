import { useQueryClient } from '@tanstack/react-query';

import type Balance from '@monetr/interface/models/Balance';
import type Spending from '@monetr/interface/models/Spending';
import type Transaction from '@monetr/interface/models/Transaction';
import request from '@monetr/interface/util/request';

export interface RemoveTransactionRequest {
  transaction: Transaction;
  adjustsBalance: boolean;
  softDelete: boolean;
}

interface RemoveTransactionResponse {
  balance: Partial<Balance>;
  spending: Partial<Spending> | null;
}

export function useRemoveTransaction(): (_: RemoveTransactionRequest) => Promise<unknown> {
  const queryClient = useQueryClient();

  // TODO It would be better to make this a mutator, however; since the response does not include the transaction that
  // was removed it is tricky to do in place mutations. At the moment this implementation instead just invalidates the
  // related queries that might be affected by the removal.
  async function removeTransaction(removal: RemoveTransactionRequest): Promise<unknown> {
    const { transaction, softDelete, adjustsBalance } = removal;
    const path = `/bank_accounts/${transaction.bankAccountId}/transactions/${transaction.transactionId}`;
    const params = new URLSearchParams();
      // Build the query parameters for the request.
      params.set('adjusts_balance', String(adjustsBalance));
      params.set('soft', String(softDelete));
    // Send the delete request to the server and handle any changes returned.
    return await request()
      .delete(`${path}?${params.toString()}`)
      .then(result => result.data)
      .then(
        async (_: RemoveTransactionResponse) =>
          await Promise.all([
            // TODO Instead of just invalidating all of these, it would be more efficient to move them into a mutator so we
            // can update their data in place.
            queryClient.invalidateQueries({
              queryKey: [`/bank_accounts/${removal.transaction.bankAccountId}/transactions`],
            }),
            queryClient.invalidateQueries({
              queryKey: [`/bank_accounts/${removal.transaction.bankAccountId}/spending`],
            }),
            queryClient.invalidateQueries({
              queryKey: [`/bank_accounts/${removal.transaction.bankAccountId}/balances`],
            }),
          ]),
      );
  }

  return removeTransaction;
}
