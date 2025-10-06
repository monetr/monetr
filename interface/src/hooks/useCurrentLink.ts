import { UseQueryResult } from '@tanstack/react-query';

import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useLink } from '@monetr/interface/hooks/useLink';
import Link from '@monetr/interface/models/Link';

export function useCurrentLink(): UseQueryResult<Link | null, unknown> {
  const { data: bankAccount } = useSelectedBankAccount();
  return useLink(bankAccount?.linkId);
}
