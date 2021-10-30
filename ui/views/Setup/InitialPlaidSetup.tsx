import React, { Component, Fragment } from "react";
import { connect } from "react-redux";

import Logo from 'assets';
import PlaidButton from "components/Plaid/PlaidButton";
import PlaidIcon from "components/Plaid/PlaidIcon";
import fetchBankAccounts from "shared/bankAccounts/actions/fetchBankAccounts";
import fetchLinks from "shared/links/actions/fetchLinks";
import logout from "shared/authentication/actions/logout";
import request from "shared/util/request";
import { getBillingEnabled } from "shared/bootstrap/selectors";

import { Button, Typography } from "@material-ui/core";
import { List } from "immutable";

interface WithConnectionPropTypes {
  billingEnabled: boolean;
  fetchBankAccounts: () => void;
  fetchLinks: () => void;
  logout: () => void;
}

interface State {
  linkId: number | null;
  loading: boolean;
  longPollAttempts: number;
}

class InitialPlaidSetup extends Component<WithConnectionPropTypes, State> {

  state = {
    linkId: null,
    loading: false,
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
        console.error(error);
        this.setState({
          loading: false,
        })
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
      .catch(error => {
        if (error.response.status === 408) {
          return this.longPollSetup();
        }
      });
  };

  onEvent = (plaidEvent: any) => {
    console.warn({
      plaidEvent,
    });
  }

  manageSubscription = () => {
    return request().get(`/billing/portal`)
      .then(result => {
        window.location.assign(result.data.url);
        return Promise.resolve();
      })
      .catch(error => {
        alert(error);
      });
  };

  renderBilling = () => {
    const { billingEnabled } = this.props;

    if (!billingEnabled) {
      return null;
    }

    return (
      <Fragment>
        <div className="w-full opacity-50 pt-2.5 pb-2.5">
          <div className="relative w-full border-t border-gray-400 top-5"/>
          <div className="relative flex justify-center inline w-full">
            <span className="relative bg-white p-1.5">
              or
            </span>
          </div>
        </div>
        <div className="w-full pt-2.5 pb-2.5">
          <Button
            onClick={ this.manageSubscription }
            color="secondary"
            className="w-full"
          >
            Manage your subscription
          </Button>
        </div>
      </Fragment>
    );
  };

  render() {
    return (
      <div className="flex justify-center w-full h-full max-h-full pb-5">
        <div
          className="flex flex-col w-full h-full p-10 xl:w-3/12 lg:w-5/12 md:w-2/3 sm:w-10/12 max-w-screen-sm sm:p-0">
          <div className="flex items-center justify-center flex-grow">
            <div>
              <div className="flex justify-center w-full mb-5">
                <img src={ Logo } className="w-1/3"/>
              </div>
              <div className="w-full pt-2.5 pb-2.5">
                <Typography
                  className="w-full text-center"
                >
                  monetr uses Plaid to retrieve bank account data automatically.
                </Typography>
              </div>
              <div className="w-full pt-2.5 pb-2.5">
                <PlaidButton
                  className="w-full"
                  color="primary"
                  disabled={ this.state.loading }
                  plaidOnSuccess={ this.onPlaidSuccess }
                  variant="contained"
                >
                  Get Started with
                  <PlaidIcon className="flex-none w-16 ml-2"/>
                </PlaidButton>
              </div>
              { this.renderBilling() }
            </div>
          </div>
          <div className="flex-initial w-full pt-2.5 pb-2.5">
            <Button
              onClick={ this.props.logout }
              className="w-full opacity-50"
            >
              Logout
            </Button>
          </div>
        </div>
      </div>
    )
  }
}

export default connect(
  state => ({
    billingEnabled: getBillingEnabled(state),
  }),
  {
    fetchBankAccounts,
    fetchLinks,
    logout,
  }
)(InitialPlaidSetup);
