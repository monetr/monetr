import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import Transaction from '@monetr/interface/models/Transaction';

export default function useSpendingTransactions(spendingId?: string): UseQueryResult<Array<Transaction>, unknown> {
  const selectedBankAccountId = useSelectedBankAccountId();
  return useQuery<Array<Partial<Transaction>>, unknown, Array<Transaction>>({
    queryKey: [`/bank_accounts/${selectedBankAccountId}/spending/${spendingId}/transactions`],
    enabled: Boolean(selectedBankAccountId) && Boolean(spendingId),
    select: data => (data ?? []).map(item => new Transaction(item)),
  });
}
