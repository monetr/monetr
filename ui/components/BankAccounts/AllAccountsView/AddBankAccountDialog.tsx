import React, { Fragment, useState } from 'react';
import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link/src/types';
import { useQueryClient } from 'react-query';
import { Close } from '@mui/icons-material';
import { Button, Dialog, DialogContent, DialogTitle, IconButton, Typography } from '@mui/material';

import AddManualBankAccountDialog from 'components/BankAccounts/AllAccountsView/AddManualBankAccountDialog';
import DuplicateInstitutionDialog from 'components/BankAccounts/AllAccountsView/DuplicateInstitutionDialog';
import PlaidButton from 'components/Plaid/PlaidButton';
import { useDetectDuplicateLink } from 'hooks/links';
import { List } from 'immutable';
import plaidLinkTokenCallback from 'shared/links/actions/plaidLinkTokenCallback';
import request from 'shared/util/request';

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

  async function longPollSetup(): Promise<void> {
    setState(prevState => ({
      longPollAttempts: prevState.longPollAttempts + 1,
    }));

    const { longPollAttempts, linkId } = state;
    if (longPollAttempts > 6) {
      return Promise.resolve();
    }

    return void request().get(`/plaid/link/setup/wait/${ linkId }`)
      .catch(error => {
        if (error.response.status === 408) {
          return this.longPollSetup();
        }

        throw error;
      });
  };

  async function afterPlaidLink(): Promise<void> {
    setState({
      duplicateDialogOpen: false,
    });
    const { callback: { token, metadata } } = state;
    return void plaidLinkTokenCallback(
      token,
      metadata.institution.institution_id,
      metadata.institution.name,
      List(metadata.accounts).map((account: { id: string }) => account.id).toArray(),
    )
      .then(result => {
        setState({
          linkId: result.linkId,
        });

        return longPollSetup()
          .then(() => Promise.all([
            queryClient.invalidateQueries('/api/links'),
            queryClient.invalidateQueries('/api/bank_accounts'),
          ]));
      })
      .catch(error => {
        setState({
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
      loading: true,
      callback: {
        token,
        metadata,
      },
    });

    if (detectDuplicateLink(metadata)) {
      setState({
        duplicateDialogOpen: true,
      });
      return Promise.resolve();
    }

    return afterPlaidLink();
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
        onConfirm={ afterPlaidLink }
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
