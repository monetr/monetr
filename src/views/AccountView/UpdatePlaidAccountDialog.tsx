import React, { Component } from "react";
import {
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography
} from "@material-ui/core";
import request from "shared/util/request";
import { PlaidConnectButton } from "views/FirstTimeSetup/PlaidConnectButton";

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

export class UpdatePlaidAccountDialog extends Component<PropTypes, State> {

  state = {
    loading: false,
    linkToken: null,
    error: null,
  };

  componentDidMount() {
    request()
      .put(`/plaid/link/update/${ this.props.linkId }`)
      .then(result => {
        this.setState({
          loading: false,
          linkToken: result.data.linkToken,
        });
      })
      .catch(error => {
        this.setState({
          loading: false,
          error: error,
        });
      });
  }

  plaidOnSuccess = (token: string, metadata: { institution: { institution_id: string, name: string }, accounts: object[] }) => {
    this.setState({
      loading: true,
    });

    request().post('/plaid/link/update/callback', {
      linkId: this.props.linkId,
      publicToken: token,
    })
      .then(result => {
        this.props.onClose();
      })
      .catch(error => {
        console.error(error);
      })
  };

  plaidOnEvent = (event: string | object) => {

  };

  plaidOnExit = (event) => {
    console.log(event);
    if (!event) {
      return;
    }

    if (event.error_code === 'item-no-error') {
      this.props.onClose();
    }
  };

  render() {
    return (
      <Dialog disableEnforceFocus={ true } open={ this.props.open } onClose={ this.props.onClose }>
        <DialogTitle>
          Update your plaid link.
        </DialogTitle>
        <DialogContent>
          <Typography>
            test
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={ this.props.onClose } color="secondary">Cancel</Button>
          { this.state.loading && <CircularProgress/> }
          { (!this.state.loading && this.state.linkToken) && <PlaidConnectButton
            token={ this.state.linkToken }
            onSuccess={ this.plaidOnSuccess }
            onExit={ this.plaidOnExit }
            onLoad={ this.plaidOnEvent }
            onEvent={ this.plaidOnEvent }
          /> }
        </DialogActions>
      </Dialog>
    );
  }
}
