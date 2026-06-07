import { useContext, useMemo } from 'react';
import { useRoute } from 'wouter';

import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import { BankAccountContext } from '@monetr/interface/components/Layout/ValidBankAccountRoute';

export function useSelectedBankAccountId(): ID<BankAccount> | null {
  const context = useContext(BankAccountContext);
  const [, params] = useRoute<{ bankId: string }>('/bank/:bankId/*');
  return useMemo(() => {
    if (params?.bankId) {
      return ID.from<BankAccount, string>(params?.bankId);
    }
    return null;
  }, [params]);
}
