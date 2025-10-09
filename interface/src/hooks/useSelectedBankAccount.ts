import { useMatch } from 'react-router-dom';
import { UseQueryResult } from '@tanstack/react-query';

import { useBankAccount } from '@monetr/interface/hooks/useBankAccount';
import BankAccount from '@monetr/interface/models/BankAccount';

export function useSelectedBankAccount(): UseQueryResult<BankAccount | undefined, unknown> {
  const match = useMatch('/bank/:bankId/*');
  const bankAccountId = match?.params?.bankId || null;

  // If we do not have a valid numeric bank account ID, but an ID was specified then something is wrong.
  if (!bankAccountId && match?.params?.bankId) {
    throw Error(`invalid bank account ID specified: "${match?.params?.bankId}" is not a valid bank account ID`);
  }

  return useBankAccount(bankAccountId);
}
