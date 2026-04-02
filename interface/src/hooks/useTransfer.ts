import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Balance from '@monetr/interface/models/Balance';
import type Spending from '@monetr/interface/models/Spending';
import request from '@monetr/interface/util/request';

export type TransferParameters = {
  fromSpendingId: string | null;
  toSpendingId: string | null;
  amount: number;
};

export function useTransfer(): (transferParameters: TransferParameters) => Promise<unknown> {
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
    return request<BalanceTransferResponse>({
      method: 'POST',
      url: `/api/bank_accounts/${selectedBankAccountId}/spending/transfer`,
      data: transferRequest,
    }).then(result => result.data);
  }

  const { mutateAsync } = useMutation({
    mutationFn: transfer,
    onSuccess: (result: BalanceTransferResponse) =>
      Promise.all([
        queryClient.setQueryData(
          [`/api/bank_accounts/${selectedBankAccountId}/spending`],
          (previous: Array<Partial<Spending>>) =>
            previous.map(item => result.spending.find(updated => updated.spendingId === item.spendingId) || item),
        ),
        result.spending.map(updatedSpending =>
          queryClient.setQueryData(
            [`/api/bank_accounts/${selectedBankAccountId}/spending/${updatedSpending.spendingId}`],
            () => updatedSpending,
          ),
        ),
        queryClient.setQueryData(
          [`/api/bank_accounts/${selectedBankAccountId}/balances`],
          (previous: Partial<Balance>) =>
            new Balance({
              ...previous,
              ...result.balance,
            }),
        ),
        queryClient.invalidateQueries({ queryKey: [`/api/bank_accounts/${selectedBankAccountId}/forecast`] }),
        queryClient.invalidateQueries({
          queryKey: [`/api/bank_accounts/${selectedBankAccountId}/forecast/next_funding`],
        }),
      ]),
  });

  return mutateAsync;
}
