import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link';
import { useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';

import { useAuthenticationSink } from './useAuthentication';

import { useBankAccounts } from 'hooks/bankAccounts';
import Link from 'models/Link';
import request from 'util/request';

export function useLinksSink(): UseQueryResult<Array<Link>> {
  const { result: { user, isActive } } = useAuthenticationSink();
  return useQuery<Array<Partial<Link>>, unknown, Array<Link>>(
    ['/links'], {
    // Only request links if there is an authenticated user.
      enabled: !!user && isActive,
      placeholderData: [],
      select: data => {
        if (Array.isArray(data)) {
          return data.map(item => new Link(item));
        }

        return [];
      },
    });
}

/**
 * @deprecated 
 */
export function useLink(linkId: number): Link | null {
  return null;
}

export function useRemoveLink(): (_linkId: number) => Promise<void> {
  const queryClient = useQueryClient();
  return async function (linkId: number): Promise<void> {
    return request()
      .delete(`/links/${linkId}`)
      .then(() => void Promise.all([
        queryClient.invalidateQueries(['/links']),
        queryClient.invalidateQueries(['/bank_accounts']),
        // TODO Invalidate other endpoints for the removed bank accounts?
      ]));
  };
}

export function useDetectDuplicateLink(): (_metadata: PlaidLinkOnSuccessMetadata) => boolean {
  const links = useLinks();
  const bankAccounts = useBankAccounts();

  return function (metadata: PlaidLinkOnSuccessMetadata): boolean {
    const linksForInstitution = new Map(Array.from(links.values())
      .filter(item => item.getIsPlaid())
      .filter(item => item.plaidInstitutionId === metadata.institution.institution_id)
      .map(item => [item.linkId, item]));

    // Check to see if the bank account we are creating is at an institution that is already added, and then check to
    // see if the mask of the account is the same. If it is then this is likely a duplicate addition.
    return Array.from(bankAccounts.values()).some(bankAccount => linksForInstitution.has(bankAccount.linkId) &&
      !!metadata.accounts.find(account => account.mask === bankAccount.mask));
  };
}

export function useTriggerManualSync(): (_linkId: number) => Promise<void> {
  const { enqueueSnackbar } = useSnackbar();
  return async (linkId: number): Promise<void> => {
    return request()
      .post('/plaid/link/sync', {
        linkId,
      })
      .then(() => void enqueueSnackbar('Triggered a manual sync in the background!', {
        variant: 'success',
        disableWindowBlurListener: true,

      }))
      .catch(error => void enqueueSnackbar(
        `Failed to trigger a manual sync: ${error?.response?.data?.error || 'unknown error'}.`,
        {
          variant: 'error',
          disableWindowBlurListener: true,
        },
      ));
  };
}
