import { useMemo } from 'react';
import { useRoute } from 'wouter';

import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';

export function useSelectedBankAccountId(): ID<BankAccount> | undefined {
  const [, params] = useRoute<{ bankId: string }>('/bank/:bankId/*');
  return useMemo(() => {
    if (params?.bankId) {
      return ID.from<BankAccount, string>(params.bankId);
    }
    return undefined;
  }, [params]);
}
