import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link';
import { useQuery, useQueryClient, UseQueryResult } from 'react-query';

import { useBankAccounts } from 'hooks/bankAccounts';
import Link from 'models/Link';
import request from 'util/request';

export type LinksResult =
  { result: Map<number, Link> }
  & UseQueryResult<Array<Partial<Link>>>;

export function useLinksSink(): LinksResult {
  const result = useQuery<Array<Partial<Link>>>('/links');
  return {
    ...result,
    result: new Map(result?.data?.map(item => {
      const link = new Link(item);
      return [link.linkId, link];
    })),
  };
}

export function useLinks(): Map<number, Link> {
  const { result } = useLinksSink();
  return result;
}

export function useLink(linkId: number): Link | null {
  const links = useLinks();
  return links.get(linkId) || null;
}

export function useRemoveLink(): (_linkId: number) => Promise<void> {
  const queryClient = useQueryClient();
  return async function (linkId: number): Promise<void> {
    return request()
      .delete(`/links/${ linkId }`)
      .then(() => void Promise.all([
        queryClient.invalidateQueries('/links'),
        queryClient.invalidateQueries('/bank_accounts'),
        // TODO Invalidate other endpoints for the removed bank accounts?
      ]));
  };
}

export function useDetectDuplicateLink(): (_metadata: PlaidLinkOnSuccessMetadata) => boolean {
  const links = useLinks();
  const bankAccounts = useBankAccounts();

  return function (metadata: PlaidLinkOnSuccessMetadata): boolean {
    const linksForInstitution = new Map(Array.from(links.values())
      .filter(item => item.plaidInstitutionId === metadata.institution.institution_id)
      .map(item => [item.linkId, item]));

    // Check to see if the bank account we are creating is at an institution that is already added, and then check to
    // see if the mask of the account is the same. If it is then this is likely a duplicate addition.
    return Array.from(bankAccounts.values()).some(bankAccount => linksForInstitution.has(bankAccount.linkId) &&
      !!metadata.accounts.find(account => account.mask === bankAccount.mask));
  };
}
