import { UseQueryResult } from '@tanstack/react-query';

import { useBankAccount } from '@monetr/interface/hooks/useBankAccount';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import BankAccount from '@monetr/interface/models/BankAccount';

export function useSelectedBankAccount(): UseQueryResult<BankAccount | undefined, unknown> {
  const bankAccountId = useSelectedBankAccountId();
  return useBankAccount(bankAccountId);
}
