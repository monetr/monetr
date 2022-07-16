import React, { useState } from 'react';
import { PlaidLinkOnEventMetadata, PlaidLinkOnExitMetadata, PlaidLinkOnSuccessMetadata } from 'react-plaid-link';
import {
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography,
} from '@mui/material';

import useMountEffect from 'hooks/useMountEffect';
import request from 'util/request';
import { PlaidConnectButton } from 'views/FirstTimeSetup/PlaidConnectButton';

interface PropTypes {
  open: boolean;
  onClose: () => void;
  linkId: number;
}

interface State {
  loading: boolean;
  linkToken: string | null;
  error: string | null;
}

export default function UpdatePlaidAccountDialog(props: PropTypes): JSX.Element {
  const [state, setState] = useState<Partial<State>>({});

  useMountEffect(() => {
    request()
      .put(`/plaid/link/update/${ props.linkId }`)
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
    })
      .then(() => props.onClose())
      .catch(error => {
        console.error(error);
      });
  }

  function plaidOnEvent(event: PlaidLinkOnEventMetadata) {

  }

  function plaidOnExit(event: PlaidLinkOnExitMetadata) {
    if (!event) {
      return;
    }

    props.onClose();
  }

  return (
    <Dialog disableEnforceFocus={ true } open={ props.open } onClose={ props.onClose }>
      <DialogTitle>
        Update your plaid link.
      </DialogTitle>
      <DialogContent>
        <Typography>
          test
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={ props.onClose } color="secondary">Cancel</Button>
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
