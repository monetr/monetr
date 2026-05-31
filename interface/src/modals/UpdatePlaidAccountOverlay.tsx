import { useCallback, useEffect, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { useQueryClient } from '@tanstack/react-query';
import {
  type PlaidLinkError,
  type PlaidLinkOnEventMetadata,
  type PlaidLinkOnExitMetadata,
  type PlaidLinkOnSuccessMetadata,
  type PlaidLinkStableEvent,
  usePlaidLink,
} from 'react-plaid-link';

import MModal from '@monetr/interface/components/MModal';
import Typography from '@monetr/interface/components/Typography';
import type Link from '@monetr/interface/models/Link';
import request from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

import styles from './UpdatePlaidAccountOverlay.module.scss';

export interface UpdatePlaidAccountOverlayProps {
  link: Link;
  updateAccountSelection?: boolean;
}

interface State {
  loading: boolean;
  linkToken: string | null;
  error: string | null;
}

function UpdatePlaidAccountOverlay({
  link,
  updateAccountSelection,
}: UpdatePlaidAccountOverlayProps): React.JSX.Element {
  const modal = useModal();
  const queryClient = useQueryClient();
  const [state, setState] = useState<Partial<State>>({
    loading: true,
    linkToken: null,
    error: null,
  });

  useEffect(() => {
    request<{ linkToken: string }>({
      method: 'PUT',
      url: `/api/plaid/link/update/${link.linkId}?update_account_selection=${!!updateAccountSelection}`,
    })
      .then(result =>
        setState({
          loading: false,
          linkToken: result.data.linkToken,
        }),
      )
      .catch(error => {
        setState({
          loading: false,
          error: error,
        });

        // TODO Add a notification that it failed.
        throw error;
      });
  }, [link, updateAccountSelection]);

  const plaidOnSuccess = useCallback(
    async (token: string, metadata: PlaidLinkOnSuccessMetadata) => {
      console.log('plaidOnSuccess', {
        token,
        metadata,
      });
      setState({
        loading: true,
      });

      return request({
        method: 'POST',
        url: '/api/plaid/link/update/callback',
        data: {
          linkId: link.linkId,
          publicToken: token,
          accountIds: metadata.accounts.map(account => account.id),
        },
      })
        .then(() =>
          Promise.all([
            queryClient.invalidateQueries({ queryKey: ['/api/bank_accounts'] }),
            queryClient.invalidateQueries({ queryKey: ['/api/links'] }),
            queryClient.invalidateQueries({ queryKey: [`/api/links/${link.linkId}`] }),
          ]),
        )
        .then(() => modal.remove());
    },
    [link, modal, queryClient],
  );

  const plaidOnExit = useCallback(
    (error: null | PlaidLinkError, metadata: PlaidLinkOnExitMetadata) => {
      console.log('plaidOnExit', {
        error,
        metadata,
      });

      modal.remove();
    },
    [modal],
  );

  const plaidOnEvent = useCallback((eventName: PlaidLinkStableEvent | string, metadata: PlaidLinkOnEventMetadata) => {
    console.log('plaidOnEvent', {
      eventName,
      metadata,
    });
  }, []);

  const { error, open } = usePlaidLink({
    token: state.linkToken ?? null,
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
    message = `One moment while we prepare Plaid to update your account selection for ${link.getName()}.`;
  } else {
    title = 'Reauthenticating';
    message = `One moment while we prepare Plaid to reauthenticate your connection to ${link.getName()}.`;
  }

  return (
    <MModal className={styles.modal} open={modal.visible}>
      <div className={styles.content}>
        <div className={styles.header}>
          <Typography className={styles.heading} size='xl' weight='bold'>
            {title}
          </Typography>
          <Typography size='lg' weight='medium'>
            {message}
          </Typography>
        </div>
      </div>
    </MModal>
  );
}

const updatePlaidAccountOverlay = NiceModal.create<UpdatePlaidAccountOverlayProps>(UpdatePlaidAccountOverlay);

export default updatePlaidAccountOverlay;

export function showUpdatePlaidAccountOverlay(props: UpdatePlaidAccountOverlayProps): Promise<void> {
  return NiceModal.show<
    void,
    ExtractProps<typeof updatePlaidAccountOverlay>,
    Partial<ExtractProps<typeof updatePlaidAccountOverlay>>
  >(updatePlaidAccountOverlay, props);
}
