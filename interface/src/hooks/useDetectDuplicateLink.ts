import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link';

import { useBankAccounts } from '@monetr/interface/hooks/useBankAccounts';
import { useLinks } from '@monetr/interface/hooks/useLinks';

export function useDetectDuplicateLink(): (_metadata: PlaidLinkOnSuccessMetadata) => boolean {
  const { data: links } = useLinks();
  const { data: bankAccounts } = useBankAccounts();

  return function (metadata: PlaidLinkOnSuccessMetadata): boolean {
    const linksForInstitution = new Map(links
      .filter(item => item.getIsPlaid())
      .filter(item => item.plaidLink?.institutionId === metadata.institution.institution_id)
      .map(item => [item.linkId, item]));

    // Check to see if the bank account we are creating is at an institution that is already added, and then check to
    // see if the mask of the account is the same. If it is then this is likely a duplicate addition.
    return Array.from(bankAccounts.values()).some(bankAccount => 
      linksForInstitution.has(bankAccount.linkId) &&
      !!metadata.accounts.find(account => account.mask === bankAccount.mask));
  };
}
