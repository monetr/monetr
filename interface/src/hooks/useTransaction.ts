import { useQuery, UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Transaction from '@monetr/interface/models/Transaction';

export function useTransaction(transactionId?: string): UseQueryResult<Transaction, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Partial<Transaction>, unknown, Transaction>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/transactions/${transactionId}`],
    enabled: Boolean(selectedBankAccountId) && Boolean(transactionId),
    select: data => new Transaction(data),
  });
}
