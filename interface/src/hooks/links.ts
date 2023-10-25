import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link';
import { useMutation, useQuery, useQueryClient, UseQueryResult } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';

import { useBankAccounts } from './bankAccounts';
import { useAuthenticationSink } from './useAuthentication';

import Link from 'models/Link';
import request from 'util/request';

export function useLinks(): UseQueryResult<Array<Link>> {
  const { result: { user, isActive } } = useAuthenticationSink();
  return useQuery<Array<Partial<Link>>, unknown, Array<Link>>(
    ['/links'], {
    // Only request links if there is an authenticated user.
      enabled: !!user && isActive,
      select: data => {
        if (Array.isArray(data)) {
          return data.map(item => new Link(item));
        }

        return [];
      },
    });
}

export function useLink(linkId: number | null): UseQueryResult<Link> {
  const queryClient = useQueryClient();
  return useQuery<Partial<Link>, unknown, Link>(
    [`/links/${linkId}`],
    {
      enabled: !!linkId,
      select: data => new Link(data),
      initialData: () => queryClient
        .getQueryData<Array<Link>>(['/links'])
        ?.find(item => item.linkId === linkId),
      initialDataUpdatedAt: () => queryClient
        .getQueryState(['/links'])?.dataUpdatedAt,
    }
  );
}

export interface CreateLinkRequest {
  institutionName: string;
  customInstitutionName: string;
  description?: string | null;
}

export function useCreateLink(): (_link: CreateLinkRequest) => Promise<Link> {
  const queryClient = useQueryClient();

  async function createLink(newLink: CreateLinkRequest): Promise<Link> {
    return request()
      .post<Partial<Link>>('/links', newLink)
      .then(result => new Link(result?.data));
  }

  const mutate = useMutation(
    createLink,
    {
      onSuccess: (newLink: Link) => Promise.all([
        queryClient.setQueriesData(
          ['/links'],
          (previous: Array<Partial<Link>> | null) => (previous ?? []).concat(newLink),
        ),
        queryClient.setQueriesData(
          [`/links/${newLink.linkId}`],
          newLink,
        ),
      ]),
    }
  );

  return mutate.mutateAsync;
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
  const { data: links } = useLinks();
  const { data: bankAccounts } = useBankAccounts();

  return function (metadata: PlaidLinkOnSuccessMetadata): boolean {
    const linksForInstitution = new Map(links
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
