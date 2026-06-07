import { useMemo } from 'react';
import { useRoute } from 'wouter';

import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';

export function useSelectedBankAccountId(): ID<BankAccount> | null {
  const [, params] = useRoute<{ bankId: string }>('/bank/:bankId/*');
  return useMemo(() => {
    return params?.bankId ? ID.from<BankAccount, string>(params?.bankId) : null;
  }, [params]);
}
