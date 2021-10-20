import React, { Component } from 'react';
import { Card, CardContent, CircularProgress, Typography } from '@material-ui/core';
import request from 'shared/util/request';
import { List } from 'immutable';
import { OAuthRedirectPlaidLink } from 'components/Plaid/OAuthRedirectPlaidLink';
import { connect } from 'react-redux';
import fetchLinks from 'shared/links/actions/fetchLinks';
import fetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import fetchBalances from 'shared/balances/actions/fetchBalances';
import { fetchFundingSchedulesIfNeeded } from 'shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded';
import fetchSpending from 'shared/spending/actions/fetchSpending';
import fetchBankAccounts from 'shared/bankAccounts/actions/fetchBankAccounts';
import { RouteComponentProps, withRouter } from 'react-router-dom';

interface WithConnectionPropTypes extends RouteComponentProps {
  fetchBalances: { (): Promise<void> }
  fetchBankAccounts: { (): Promise<void> }
  fetchFundingSchedulesIfNeeded: { (): Promise<void> }
  fetchInitialTransactionsIfNeeded: { (): Promise<void> }
  fetchLinks: { (): Promise<void> }
  fetchSpending: { (): Promise<void> }
}

interface State {
  loading: boolean;
  linkToken: string;
  error: string | null;
  linkId: number | null;
  longPollAttempts: number;
}

class OAuthRedirect extends Component<WithConnectionPropTypes, State> {

  state = {
    loading: true,
    linkToken: '',
    error: null,
    linkId: null,
    longPollAttempts: 0,
  };

  componentDidMount() {
    request()
      .get('/plaid/link/token/new?use_cache=true')
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
        })
      });
  }

  plaidLinkSuccess = (token: string, metadata: any) => {
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
              this.props.fetchBankAccounts().then(() => {
                return Promise.all([
                  this.props.fetchInitialTransactionsIfNeeded(),
                  this.props.fetchFundingSchedulesIfNeeded(),
                  this.props.fetchSpending(),
                  this.props.fetchBalances(),
                ]);
              }),
            ]);
          });
      })
      .catch(error => {
        console.error(error);
      })
  };

  plaidLinkExit = (error: object, metadata: object) => {
    console.error({
      error,
      metadata,
    });
    this.props.history.push('/');
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

  renderLoading = () => {
    return (
      <div>
        <Typography variant="h5">
          One moment...
        </Typography>
        <div className="flex justify-center items-center p-5 m-5">
          <CircularProgress/>
        </div>
      </div>
    );
  };

  onEvent = (event, metadata) => {
    console.warn({
      event,
      metadata,
    });
  }

  renderReady = () => {
    const { linkToken } = this.state;

    return (
      <div>
        <OAuthRedirectPlaidLink
          linkToken={ linkToken }
          plaidOnSuccess={ this.plaidLinkSuccess }
          plaidOnExit={ this.plaidLinkExit }
        />
      </div>
    );
  };

  render() {
    return (
      <div className="w-full h-full flex justify-center items-center p-10">
        <div>
          <Card>
            <CardContent>
              { this.renderLoading() }
              { !this.state.loading && this.renderReady() }
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }
}

export default connect(
  state => ({}),
  {
    fetchBalances,
    fetchBankAccounts,
    fetchFundingSchedulesIfNeeded,
    fetchInitialTransactionsIfNeeded,
    fetchLinks,
    fetchSpending,
  },
)(withRouter(OAuthRedirect));
