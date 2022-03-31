import Link from 'models/Link';
import React, { Component, Fragment } from 'react';
import { Button, Dialog, DialogContent, DialogTitle, IconButton, Typography } from '@mui/material';
import { PlaidLinkOnSuccessMetadata } from 'react-plaid-link/src/types/index';
import { connect } from 'react-redux';
import { Close } from '@mui/icons-material';
import PlaidButton from 'components/Plaid/PlaidButton';
import { List } from 'immutable';
import detectDuplicateLink from 'shared/links/actions/detectDuplicateLink';
import plaidLinkTokenCallback from 'shared/links/actions/plaidLinkTokenCallback';
import { getLinksByInstitutionId } from 'shared/links/selectors/getLinksByInstitutionId';
import request from 'shared/util/request';
import fetchBankAccounts from 'shared/bankAccounts/actions/fetchBankAccounts';
import fetchLinks from 'shared/links/actions/fetchLinks';
import { AppState } from 'store';
import AddManualBankAccountDialog from 'components/BankAccounts/AllAccountsView/AddManualBankAccountDialog';
import { Map } from 'immutable';
import DuplicateInstitutionDialog from 'components/BankAccounts/AllAccountsView/DuplicateInstitutionDialog';

export interface PropTypes {
  open: boolean;
  onClose: () => void;
}

interface WithConnectionPropTypes extends PropTypes {
  fetchLinks: () => Promise<void>;
  fetchBankAccounts: () => Promise<void>;
  detectDuplicateLink: (metadata: PlaidLinkOnSuccessMetadata) => boolean;
  linksByInstitutionId: Map<string, Link[]>;
}

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

class AddBankAccountDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: false,
    linkId: null,
    longPollAttempts: 0,
    manualDialogOpen: false,
    duplicateDialogOpen: false,
    callback: null,
  };

  onPlaidSuccess = (token: string, metadata: PlaidLinkOnSuccessMetadata): Promise<void> => {
    const { detectDuplicateLink } = this.props;
    this.setState({
      loading: true,
      callback: {
        token,
        metadata,
      },
    });

    if (detectDuplicateLink(metadata)) {
      this.setState({
        duplicateDialogOpen: true,
      });
      return Promise.resolve();
    }

    return this.afterPlaidLink();
  }

  afterPlaidLink = () => {
    this.setState({
      duplicateDialogOpen: false,
    })
    const { callback: { token, metadata } } = this.state;
    return plaidLinkTokenCallback(
      token,
      metadata.institution.institution_id,
      metadata.institution.name,
      List(metadata.accounts).map((account: { id: string }) => account.id).toArray(),
    )
      .then(result => {
        this.setState({
          linkId: result.linkId,
        });

        return this.longPollSetup()
          .then(() => {
            return this.props.fetchLinks().then(() => this.props.fetchBankAccounts());
          });
      })
      .catch(error => {
        this.setState({
          loading: false,
        })
      })
      .finally(() => {
        this.props.onClose();
      });
  };

  longPollSetup = () => {
    this.setState(prevState => ({
      longPollAttempts: prevState.longPollAttempts + 1,
    }));

    const { longPollAttempts, linkId } = this.state;
    if (longPollAttempts > 6) {
      return Promise.resolve();
    }

    return request().get(`/plaid/link/setup/wait/${ linkId }`)
      .then(result => {
        return Promise.resolve();
      })
      .catch(error => {
        if (error.response.status === 408) {
          return this.longPollSetup();
        }
      });
  };

  openManualDialog = () => this.setState({
    manualDialogOpen: true,
  });

  closeManualDialog = () => this.setState({
    manualDialogOpen: false,
  });

  renderDialogs = () => {
    const { manualDialogOpen, duplicateDialogOpen } = this.state;

    if (manualDialogOpen) {
      return <AddManualBankAccountDialog open={ true } onClose={ this.closeManualDialog }/>
    }

    if (duplicateDialogOpen) {
      return <DuplicateInstitutionDialog
        open={ true }
        onCancel={ this.props.onClose }
        onConfirm={ this.afterPlaidLink }
      />
    }

    return null;
  }

  render() {
    const { open, onClose } = this.props;

    return (
      <Fragment>

        { this.renderDialogs() }

        <Dialog open={ open } disableEnforceFocus={ true } maxWidth="xs">
          <DialogTitle>
            <div className="flex items-center">
              <span className="text-2xl flex-auto">
                Add a bank account
              </span>
              <IconButton className="flex-none" onClick={ onClose }>
                <Close/>
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
                disabled={ this.state.loading }
                useCache={ true }
                plaidOnSuccess={ this.onPlaidSuccess }
                variant="outlined"
                color="primary"
              >
                Connect My Bank Account
              </PlaidButton>
              <Button
                variant="outlined"
                onClick={ this.openManualDialog }
              >
                Create Manual Bank Account
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </Fragment>
    );
  }
}

export default connect(
  (state: AppState) => ({
    linksByInstitutionId: getLinksByInstitutionId(state),
  }),
  {
    fetchLinks,
    fetchBankAccounts,
    detectDuplicateLink,
  }
)(AddBankAccountDialog);
