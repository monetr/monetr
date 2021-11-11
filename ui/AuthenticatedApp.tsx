import NavigationBar from 'NavigationBar';
import React, { Component, Fragment } from 'react';
import { getHasAnyLinks } from 'shared/links/selectors/getHasAnyLinks';
import fetchBalances from 'shared/balances/actions/fetchBalances';
import fetchBankAccounts from 'shared/bankAccounts/actions/fetchBankAccounts';
import { fetchFundingSchedulesIfNeeded } from 'shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded';
import fetchSpending from 'shared/spending/actions/fetchSpending';
import fetchLinksIfNeeded from 'shared/links/actions/fetchLinksIfNeeded';
import fetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import { Redirect, Route, RouteComponentProps, Switch, withRouter } from 'react-router-dom';
import { connect } from 'react-redux';
import { Backdrop, CircularProgress } from '@mui/material';
import { AppState } from 'store';
import TransactionsView from 'views/Transactions/TransactionsView';
import ExpensesView from 'views/Expenses/ExpensesView';
import GoalsView from 'views/Goals/GoalsView';
import OAuthRedirect from 'views/FirstTimeSetup/OAuthRedirect';
import AllAccountsView from 'views/AccountView/AllAccountsView';
import Logout from 'views/Authentication/Logout';
import InitialPlaidSetup from 'views/Setup/InitialPlaidSetup';

interface WithConnectionPropTypes {
  fetchBalances: () => Promise<any>;
  fetchBankAccounts: () => Promise<any>;
  fetchFundingSchedulesIfNeeded: () => Promise<any>;
  fetchInitialTransactionsIfNeeded: () => Promise<any>;
  fetchLinksIfNeeded: () => Promise<any>;
  fetchSpending: () => Promise<any>;
  hasAnyLinks: boolean;
}

interface State {
  loading: boolean;
}

export class AuthenticatedApp extends Component<RouteComponentProps & WithConnectionPropTypes, State> {

  state = {
    loading: true,
  };

  componentDidMount() {
    const {
      fetchBalances,
      fetchBankAccounts,
      fetchFundingSchedulesIfNeeded,
      fetchSpending,
      fetchLinksIfNeeded,
      fetchInitialTransactionsIfNeeded,
    } = this.props;

    Promise.all([
      fetchLinksIfNeeded(),
      fetchBankAccounts().then(() => Promise.all([
        fetchInitialTransactionsIfNeeded(),
        fetchFundingSchedulesIfNeeded(),
        fetchSpending(),
        fetchBalances(),
      ])),
    ])
      .finally(() => this.setState({ loading: false }));
  }

  renderSubRoutes = () => {
    if (this.props.hasAnyLinks) {
      return this.renderSetup();
    }

    return this.renderNotSetup()
  };

  renderNotSetup = () => {
    return (
      <Switch>
        <Route path="/logout" exact component={ Logout }/>
        <Route path="/setup" exact component={ InitialPlaidSetup }/>
        <Route path="/plaid/oauth-return">
          <OAuthRedirect/>
        </Route>
        <Route path="/">
          <Redirect to="/setup"/>
        </Route>
        <Route>
          <Redirect to="/setup"/>
        </Route>
      </Switch>
    )
  };

  renderSetup = () => {
    return (
      <Fragment>
        <NavigationBar/>
        <Switch>
          <Route path="/register">
            <Redirect to="/"/>
          </Route>
          <Route path="/login">
            <Redirect to="/"/>
          </Route>
          <Route path="/logout" exact component={ Logout }/>
          <Route path="/transactions" exact component={ TransactionsView }/>
          <Route path="/expenses" exact component={ ExpensesView }/>
          <Route path="/goals" exact component={ GoalsView }/>
          <Route path="/accounts" exact component={ AllAccountsView }/>
          <Route path="/">
            <Redirect to="/transactions"/>
          </Route>
          <Route>
            <h1>Not found</h1>
          </Route>
        </Switch>
      </Fragment>
    );
  }

  render() {
    if (this.state.loading) {
      return (
        <Backdrop open={ true }>
          <CircularProgress color="inherit"/>
        </Backdrop>
      );
    }

    return this.renderSubRoutes();
  }
}

export default connect(
  (state: AppState) => ({
    hasAnyLinks: getHasAnyLinks(state),
  }),
  {
    fetchBalances,
    fetchBankAccounts,
    fetchFundingSchedulesIfNeeded,
    fetchSpending,
    fetchLinksIfNeeded,
    fetchInitialTransactionsIfNeeded,
  }
)(withRouter(AuthenticatedApp));
