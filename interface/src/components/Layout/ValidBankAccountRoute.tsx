import { createContext, useMemo } from 'react';
import { useRoute } from 'wouter';

import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';

export interface BankAccountContextInvalid {
  bankAccountId: null;
  valid: false;
}

export interface BankAccountContextValid {
  bankAccountId: ID<BankAccount>;
  valid: true;
}

export type IBankAccountContext = BankAccountContextValid | BankAccountContextInvalid;

export const BankAccountContext = createContext<IBankAccountContext>({
  bankAccountId: null,
  valid: false,
});

export default function ValidBankAccountRoute(props: React.PropsWithChildren): React.ReactNode {
  const [, params] = useRoute<{ bankId: string }>('/bank/:bankId/*');
  const bankAccountId = useMemo(() => {
    if (params?.bankId) {
      return ID.from<BankAccount, string>(params?.bankId);
    }
    return null;
  }, [params]);

  if (bankAccountId) {
    return (
      <BankAccountContext.Provider value={{ bankAccountId, valid: true }}>{props.children}</BankAccountContext.Provider>
    );
  }

  return null;
}
