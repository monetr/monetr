import { useMutation, useQueryClient } from '@tanstack/react-query';

import Balance from '@monetr/interface/models/Balance';
import type Spending from '@monetr/interface/models/Spending';
import type Transaction from '@monetr/interface/models/Transaction';
import type { WithJsonValues } from '@monetr/interface/util/json';
import request, { type ApiError } from '@monetr/interface/util/request';

export interface CreateTransactionRequest {
  name: string;
  amount: number;
  spendingId: string | null;
  date: Date;
  merchantName: string | null;
  isPending: boolean;

  // These are auxilary fields
  bankAccountId: string;
  adjustsBalance: boolean;
}

export interface CreateTransactionResponse {
  transaction: Partial<Transaction>;
  balance: WithJsonValues<Balance>;
  spending: Partial<Spending> | null;
}

export type CreateTransactionError =
  | { error: string; problems: never }
  | { error: string; problems: { [K in keyof CreateTransactionRequest]: string } };

export function useCreateTransaction(): (
  _: CreateTransactionRequest,
) => Promise<CreateTransactionResponse | CreateTransactionError> {
  const queryClient = useQueryClient();

  async function createTransaction({
    // The bank account Id field is dropped by the controller, the path value is authoritative here so even thoguh this
    // does kind of betray the typing of [CreateTransactionError] this is correct.
    bankAccountId,
    ...transaction
  }: CreateTransactionRequest): Promise<CreateTransactionResponse> {
    return await request<CreateTransactionResponse>({
      method: 'POST',
      url: `/api/bank_accounts/${bankAccountId}/transactions`,
      data: transaction,
    })
      .then(result => result.data)
      .catch((error: ApiError<CreateTransactionError>) => {
        throw error.response.data;
      });
  }

  const { mutateAsync } = useMutation({
    mutationFn: createTransaction,
    onSuccess: ({ transaction, balance, spending }: CreateTransactionResponse) =>
      Promise.all([
        queryClient.setQueryData(
          [`/api/bank_accounts/${transaction.bankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            previous.map(item => (spending?.spendingId === item.spendingId ? spending : item)),
        ),
        spending != null &&
          queryClient.setQueryData(
            [`/api/bank_accounts/${transaction.bankAccountId}/spending/${spending.spendingId}`],
            spending,
          ),
        queryClient.setQueryData(
          [`/api/bank_accounts/${transaction.bankAccountId}/balances`],
          (previous: Partial<Balance>) =>
            new Balance({
              ...previous,
              ...balance,
            }),
        ),
        queryClient.invalidateQueries({
          queryKey: [`/api/bank_accounts/${transaction.bankAccountId}/transactions`],
        }),
      ]),
  });

  return mutateAsync;
}
