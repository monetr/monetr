import React, { Fragment, useState } from 'react';
import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link/src/types';
import { useQueryClient } from 'react-query';
import { Close } from '@mui/icons-material';
import { Button, Dialog, DialogContent, DialogTitle, IconButton, Typography } from '@mui/material';

import AddManualBankAccountDialog from 'components/BankAccounts/AllAccountsView/AddManualBankAccountDialog';
import DuplicateInstitutionDialog from 'components/BankAccounts/AllAccountsView/DuplicateInstitutionDialog';
import PlaidButton from 'components/Plaid/PlaidButton';
import { useDetectDuplicateLink } from 'hooks/links';
import plaidLinkTokenCallback from 'util/plaidLinkTokenCallback';
import request from 'util/request';

interface State {
  loading: boolean;
  linkId: number | null;
  longPollAttempts: number;
  manualDialogOpen: boolean;
  duplicateDialogOpen: boolean;
  callback: {
    token: string;
    metadata: PlaidLinkOnSuccessMetadata;
  } | null;
}

interface Props {
  open: boolean;
  onClose: () => void;
}

export default function AddBankAccountDialog(props: Props): JSX.Element {
  const queryClient = useQueryClient();
  const detectDuplicateLink = useDetectDuplicateLink();

  const [state, setState] = useState<Partial<State>>({});

  async function longPollSetup(linkId: number, attempts: number = 0): Promise<void> {
    if (attempts > 6) {
      return Promise.resolve();
    }

    return void request().get(`/plaid/link/setup/wait/${ linkId }`)
      .catch(error => {
        if (error.response.status === 408) {
          return longPollSetup(linkId, attempts + 1);
        }

        throw error;
      });
  };

  async function afterPlaidLink(token: string, metadata: PlaidLinkOnSuccessMetadata): Promise<void> {
    setState({
      ...state,
      duplicateDialogOpen: false,
    });
    return void plaidLinkTokenCallback(
      token,
      metadata.institution.institution_id,
      metadata.institution.name,
      metadata.accounts.map((account: { id: string }) => account.id),
    )
      .then(async result => {
        return longPollSetup(result.linkId)
          .then(() => Promise.all([
            queryClient.invalidateQueries('/links'),
            queryClient.invalidateQueries('/bank_accounts'),
          ]));
      })
      .catch(error => {
        setState({
          ...state,
          loading: false,
        });

        throw error;
      })
      .finally(() => {
        props.onClose();
      });
  }

  async function onPlaidSuccess(token: string, metadata: PlaidLinkOnSuccessMetadata): Promise<void> {
    setState({
      ...state,
      loading: true,
      callback: {
        token,
        metadata,
      },
    });

    if (detectDuplicateLink(metadata)) {
      setState({
        ...state,
        duplicateDialogOpen: true,
      });
      return Promise.resolve();
    }

    return afterPlaidLink(token, metadata);
  }

  const openManualDialog = () => setState({
    manualDialogOpen: true,
  });

  const closeManualDialog = () => setState({
    manualDialogOpen: false,
  });

  function Dialogs(): JSX.Element {
    const { manualDialogOpen, duplicateDialogOpen } = state;

    if (manualDialogOpen) {
      return <AddManualBankAccountDialog open={ true } onClose={ closeManualDialog } />;
    }

    if (duplicateDialogOpen) {
      return <DuplicateInstitutionDialog
        open={ true }
        onCancel={ props.onClose }
        onConfirm={ () => alert('TODO') }
      />;
    }

    return null;
  }

  const { open, onClose } = props;
  return (
    <Fragment>
      <Dialogs />
      <Dialog open={ open } disableEnforceFocus={ true } maxWidth="xs">
        <DialogTitle>
          <div className="flex items-center">
            <span className="text-2xl flex-auto">
                Add a bank account
            </span>
            <IconButton className="flex-none" onClick={ onClose }>
              <Close />
            </IconButton>
          </div>
        </DialogTitle>
        <DialogContent className="mb-5">
          <Typography>
            You can link your bank account to automatically sync transactions and balances. Or you can create a
            manual bank account to manage your transactions and balances yourself.
          </Typography>
          <div className="grid grid-flow-col grid-rows-2 grid-cols-1 gap-2 mt-5">
            <PlaidButton
              disabled={ state.loading }
              useCache={ true }
              plaidOnSuccess={ onPlaidSuccess }
              variant="outlined"
              color="primary"
            >
              Connect My Bank Account
            </PlaidButton>
            <Button
              variant="outlined"
              onClick={ openManualDialog }
            >
              Create Manual Bank Account
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </Fragment>
  );
}
