import React, { Component, Fragment } from "react";
import { Button, Dialog, DialogContent, DialogTitle, IconButton, Typography } from "@material-ui/core";
import { connect } from "react-redux";
import { Close } from "@material-ui/icons";
import PlaidButton from "Plaid/PlaidButton";
import { List } from "immutable";
import request from "shared/util/request";
import fetchBankAccounts from "shared/bankAccounts/actions/fetchBankAccounts";
import fetchLinks from "shared/links/actions/fetchLinks";

export interface PropTypes {
  open: boolean;
  onClose: () => void;
}

interface WithConnectionPropTypes extends PropTypes {
  fetchLinks: () => void;
  fetchBankAccounts: () => void;
}

interface State {
  loading: boolean;
  linkId: number | null;
  longPollAttempts: number;
}

class AddBankAccountDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: false,
    linkId: null,
    longPollAttempts: 0,
  };

  onPlaidSuccess = (token: string, metadata: any) => {
    this.setState({
      loading: true,
    });

    return request().post('/plaid/link/token/callback', {
      publicToken: token,
      institutionId: metadata.institution.institution_id,
      institutionName: metadata.institution.name,
      accountIds: List(metadata.accounts).map((account: { id: string }) => account.id).toArray()
    })
      .then(result => {
        this.setState({
          linkId: result.data.linkId,
        });

        return this.longPollSetup()
          .then(() => {
            return Promise.all([
              this.props.fetchLinks(),
              this.props.fetchBankAccounts(),
            ]);
          });
      })
      .catch(error => {
        alert(error);
      })
      .finally(() => {
        this.props.onClose();
      });
  }

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

  onEvent = (thing) => {
    console.warn({
      thing,
    });
  }

  render() {
    const { open, onClose } = this.props;

    return (
      <Fragment>
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
                onSuccess={ this.onPlaidSuccess }
                onExit={ this.onEvent }
                onLoad={ this.onEvent }
                onEvent={ this.onEvent }
                variant="outlined"
                color="primary"
              >
                Connect My Bank Account
              </PlaidButton>
              <Button variant="outlined">
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
  state => ({}),
  {
    fetchLinks,
    fetchBankAccounts,
  }
)(AddBankAccountDialog);
