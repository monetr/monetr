import React, { useState } from 'react';
import { PlaidLinkOnEventMetadata, PlaidLinkOnExitMetadata, PlaidLinkOnSuccessMetadata } from 'react-plaid-link';
import { useQueryClient } from 'react-query';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import {
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography,
} from '@mui/material';
import * as Sentry from '@sentry/react';

import useMountEffect from 'hooks/useMountEffect';
import request from 'util/request';
import { PlaidConnectButton } from 'views/FirstTimeSetup/PlaidConnectButton';

interface UpdatePlaidAccountDialogProps {
  linkId: number;
  updateAccountSelection?: boolean;
}

interface State {
  loading: boolean;
  linkToken: string | null;
  error: string | null;
}

function UpdatePlaidAccountDialog(props: UpdatePlaidAccountDialogProps): JSX.Element {
  const modal = useModal();
  const queryClient = useQueryClient();
  const [state, setState] = useState<Partial<State>>({});

  useMountEffect(() => {
    request()
      .put(`/plaid/link/update/${ props.linkId }?update_account_selection=${ !!props.updateAccountSelection }`)
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
  });

  async function plaidOnSuccess(token: string, metadata: PlaidLinkOnSuccessMetadata) {
    setState({
      loading: true,
    });

    request().post('/plaid/link/update/callback', {
      linkId: props.linkId,
      publicToken: token,
      accountIds: metadata.accounts.map(account => account.id),
    })
      .then(() => queryClient.invalidateQueries('/bank_accounts'))
      .then(() => modal.remove())
      .catch(error => {
        // If sentry is configured I want to know when errors happen, mostly because the changes I'm making right now
        // are to reduce errors. But I need visibility into whether or not it actually does anything.
        Sentry.captureException(error);
        console.error(error);
      });
  }

  function plaidOnEvent(_event: PlaidLinkOnEventMetadata) {
    return;
  }

  function plaidOnExit(event: PlaidLinkOnExitMetadata) {
    if (!event) {
      return;
    }

    modal.remove();
  }

  return (
    <Dialog
      disableEnforceFocus={ true }
      open={ modal.visible }
      onClose={ modal.remove }
    >
      <DialogTitle>
        Update your plaid link.
      </DialogTitle>
      <DialogContent>
        <Typography>
          One moment...
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button
          onClick={ modal.remove }
          color="secondary"
        >
          Cancel
        </Button>
        { state.loading && <CircularProgress /> }
        { (!state.loading && state.linkToken) && <PlaidConnectButton
          token={ state.linkToken }
          onSuccess={ plaidOnSuccess }
          onExit={ plaidOnExit }
          onLoad={ plaidOnEvent }
          onEvent={ plaidOnEvent }
        /> }
      </DialogActions>
    </Dialog>
  );
}

const updatePlaidAccountModal = NiceModal.create(UpdatePlaidAccountDialog);
export default updatePlaidAccountModal;

export function showUpdatePlaidAccountDialog(props: UpdatePlaidAccountDialogProps): void {
  NiceModal.show(updatePlaidAccountModal, props);
}
