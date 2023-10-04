import React, { useCallback, useEffect, useState } from 'react';
import { PlaidLinkError, PlaidLinkOnEventMetadata, PlaidLinkOnExitMetadata, PlaidLinkOnSuccessMetadata, PlaidLinkStableEvent, usePlaidLink } from 'react-plaid-link';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { useQueryClient } from '@tanstack/react-query';

import MModal from 'components/MModal';
import MSpan from 'components/MSpan';
import Link from 'models/Link';
import request from 'util/request';
import { ExtractProps } from 'util/typescriptEvils';

export interface UpdatePlaidAccountOverlayProps {
  link: Link;
  updateAccountSelection?: boolean;
}

interface State {
  loading: boolean;
  linkToken: string | null;
  error: string | null;
}

function UpdatePlaidAccountOverlay({ link, updateAccountSelection }: UpdatePlaidAccountOverlayProps): JSX.Element {
  const modal = useModal();
  const queryClient = useQueryClient();
  const [state, setState] = useState<Partial<State>>({
    loading: true,
    linkToken: null,
    error: null,
  });

  useEffect(() => {
    request()
      .put(`/plaid/link/update/${ link.linkId }?update_account_selection=${ !!updateAccountSelection }`)
      .then(result => setState({
        loading: false,
        linkToken: result.data.linkToken,
      }))
      .catch(error => {
        setState({
          loading: false,
          error: error,
        });

        // TODO Add a notification that it failed.
        throw error;
      });
  }, [link, updateAccountSelection]);


  const plaidOnSuccess = useCallback(async (token: string, metadata: PlaidLinkOnSuccessMetadata) => {
    console.log('plaidOnSuccess', {
      token,
      metadata,
    });
    setState({
      loading: true,
    });

    return request().post('/plaid/link/update/callback', {
      linkId: link.linkId,
      publicToken: token,
      accountIds: metadata.accounts.map(account => account.id),
    })
      .then(() => Promise.all([
        queryClient.invalidateQueries(['/bank_accounts']),
        queryClient.invalidateQueries(['/links']),
        queryClient.invalidateQueries([`/links/${link.linkId}`]),
      ]))
      .then(() => modal.remove());
  }, [link, modal, queryClient]);

  const plaidOnExit = useCallback((error: null | PlaidLinkError, metadata: PlaidLinkOnExitMetadata) => {
    console.log('plaidOnExit', {
      error,
      metadata,
    });
    if (!metadata || !error) return;

    modal.remove();
  }, [modal]);

  const plaidOnEvent = useCallback((eventName: PlaidLinkStableEvent | string, metadata: PlaidLinkOnEventMetadata) => {
    console.log('plaidOnEvent', {
      eventName,
      metadata,
    });
  }, []);

  const { error, open } = usePlaidLink({
    token: state.linkToken,
    onSuccess: plaidOnSuccess,
    onExit: plaidOnExit,
    onEvent: plaidOnEvent,
  });

  useEffect(() => {
    if (error) {
      console.error('PLAID LINK ERROR', error);
      return;
    }

    if (open && state.linkToken) {
      open();
    }
  }, [error, open, state.linkToken]);

  let title: string, message: string;
  if (updateAccountSelection) {
    title = 'Updating Account Selection';
    message = `One moment while we prepare Plaid to update your account selection for ${ link.getName() }.`;
  } else {
    title = 'Reauthenticating';
    message = `One moment while we prepare Plaid to reauthenticate your connection to ${link.getName()}.`;
  }

  return (
    <MModal open={ modal.visible } className='py-4 md:max-w-md'>
      <div className='h-full flex flex-col gap-4 p-2 justify-between'>
        <div className='flex flex-col'>
          <MSpan weight='bold' size='xl' className='mb-2'>
            { title }
          </MSpan>
          <MSpan size='lg' weight='medium'>
            { message }
          </MSpan>
        </div>
      </div>
    </MModal>
  );
}

const updatePlaidAccountOverlay = NiceModal.create<UpdatePlaidAccountOverlayProps>(UpdatePlaidAccountOverlay);

export default updatePlaidAccountOverlay;

export function showUpdatePlaidAccountOverlay(props: UpdatePlaidAccountOverlayProps): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof updatePlaidAccountOverlay>, {}>(updatePlaidAccountOverlay, props);
}
