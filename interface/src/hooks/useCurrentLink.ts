import { UseQueryResult } from '@tanstack/react-query';

import { useLink } from '@monetr/interface/hooks/useLink';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import Link from '@monetr/interface/models/Link';

export function useCurrentLink(): UseQueryResult<Link | null, unknown> {
  const { data: bankAccount } = useSelectedBankAccount();
  return useLink(bankAccount?.linkId);
}
