import { useMutation, useQueryClient } from '@tanstack/react-query';

import Balance from '@monetr/interface/models/Balance';
import Spending from '@monetr/interface/models/Spending';
import Transaction from '@monetr/interface/models/Transaction';
import request from '@monetr/interface/util/request';

export interface CreateTransactionRequest {
  name: string;
  bankAccountId: string;
  amount: number;
  spendingId: string | null;
  date: Date;
  merchantName: string | null;
  isPending: boolean;
  adjustsBalance: boolean;
}

export interface CreateTransactionResponse {
  transaction: Partial<Transaction>;
  balance: Partial<Balance>;
  spending: Partial<Spending> | null;
}

export function useCreateTransaction(): (_: CreateTransactionRequest) => Promise<CreateTransactionResponse> {
  const queryClient = useQueryClient();

  async function createTransaction(transaction: CreateTransactionRequest): Promise<CreateTransactionResponse> {
    return request()
      .post<CreateTransactionResponse>(`/bank_accounts/${transaction.bankAccountId}/transactions`, transaction)
      .then(result => result.data);
  }

  const { mutateAsync } = useMutation({
    mutationFn: createTransaction,
    onSuccess: ({ transaction, balance, spending }: CreateTransactionResponse) =>
      Promise.all([
        queryClient.setQueryData(
          [`/bank_accounts/${transaction.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            previous.map(item => (spending?.spendingId == item.spendingId ? spending : item)),
        ),
        spending != null &&
          queryClient.setQueryData(
            [`/bank_accounts/${transaction.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          ),
        queryClient.setQueryData(
          [`/bank_accounts/${transaction.bankAccountId}/balances`],
          (previous: Partial<Balance>) =>
            new Balance({
              ...previous,
              ...balance,
            }),
        ),
        queryClient.invalidateQueries({
          queryKey: [`/bank_accounts/${transaction.bankAccountId}/transactions`],
        }),
      ]),
  });

  return mutateAsync;
}
