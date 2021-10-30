import React, { Component } from "react";
import request from "shared/util/request";
import {
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography
} from "@mui/material";
import { PlaidConnectButton } from "views/FirstTimeSetup/PlaidConnectButton";
import { List } from "immutable";
import { connect } from "react-redux";
import fetchLinks from "shared/links/actions/fetchLinks";
import fetchBankAccounts from "shared/bankAccounts/actions/fetchBankAccounts";

export interface Props {
  open: boolean;
  onClose: () => void;
}

interface WithConnectionProps extends Props {
  fetchLinks: () => Promise<void>;
  fetchBankAccounts: () => Promise<void>;
}

interface State {
  loading: boolean;
  linkToken: string | null;
  error: string | null;
  linkId: number | null;
  longPollAttempts: number;
}

export class AddPlaidDialog extends Component<WithConnectionProps, State> {

  state = {
    loading: true,
    linkToken: null,
    error: null,
    linkId: null,
    longPollAttempts: 0
  };

  componentDidMount() {
    request()
      .get(`/plaid/link/token/new`)
      .then(result => {
        this.setState({
          loading: false,
          linkToken: result.data.linkToken,
        })
      })
      .catch(error => {
        this.setState({
          loading: false,
          error: error,
        });
      })
  }

  plaidLinkSuccess = (token, metadata) => {
    this.setState({
      loading: true,
    });

    request().post('/plaid/link/token/callback', {
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
        console.error(error);
      })
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

  onEvent = (thing, stuff) => {
    console.warn({
      thing,
      stuff
    });
  }

  renderPlaidButton = () => {
    const { loading, linkToken } = this.state;

    if (loading) {
      return <CircularProgress/>
    }

    if (linkToken) {
      return <PlaidConnectButton
        token={ linkToken }
        onSuccess={ this.plaidLinkSuccess }
        onEvent={ this.onEvent }
        onExit={ this.onEvent }
        onLoad={ this.onEvent }
      />
    }

    return <Typography>Something went wrong...</Typography>
  };

  render() {
    return (
      <Dialog disableEnforceFocus={ true } open={ this.props.open } onClose={ this.props.onClose }>
        <DialogTitle>
          Add another Plaid link?
        </DialogTitle>
        <DialogContent>
          <Typography>
            You can add additional Plaid links to your account. This will add bank account's available to help with
            budgeting.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={ this.props.onClose }>Cancel</Button>
          { this.renderPlaidButton() }
        </DialogActions>
      </Dialog>
    )
  }
}

export default connect(
  state => ({}),
  {
    fetchLinks,
    fetchBankAccounts,
  }
)(AddPlaidDialog);
