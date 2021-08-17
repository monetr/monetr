import React, { Component, Fragment } from "react";
import { getHasAnyLinks } from "shared/links/selectors/getHasAnyLinks";
import logout from "shared/authentication/actions/logout";
import fetchBalances from "shared/balances/actions/fetchBalances";
import fetchBankAccounts from "shared/bankAccounts/actions/fetchBankAccounts";
import { fetchFundingSchedulesIfNeeded } from "shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded";
import fetchSpending from "shared/spending/actions/fetchSpending";
import fetchLinksIfNeeded from "shared/links/actions/fetchLinksIfNeeded";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";
import { Link as RouterLink, Redirect, Route, RouteComponentProps, Switch, withRouter } from "react-router-dom";
import { connect } from "react-redux";
import { AppBar, Backdrop, Button, CircularProgress, IconButton, Menu, MenuItem, Toolbar } from "@material-ui/core";
import BankAccountSelector from "components/BankAccounts/BankAccountSelector";
import BalanceNavDisplay from "components/Balance/BalanceNavDisplay";
import MenuIcon from "@material-ui/icons/Menu";
import TransactionsView from "views/TransactionsView";
import ExpensesView from "views/ExpensesView";
import GoalsView from "views/GoalsView";
import AccountView from "views/AccountView";
import OAuthRedirect from "views/FirstTimeSetup/OAuthRedirect";
import AllAccountsView from "views/AccountView/AllAccountsView";
import AfterCheckout from "views/Subscriptions/AfterCheckout";
import Logout from "views/Authentication/Logout";
import InitialPlaidSetup from "views/Setup/InitialPlaidSetup";
import request from "shared/util/request";
import { CreditCard, ExitToApp } from "@material-ui/icons";

interface WithConnectionPropTypes {
  logout: () => void;
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
  menuAnchorEl: Element | null;
}

export class AuthenticatedApp extends Component<RouteComponentProps & WithConnectionPropTypes, State> {

  state = {
    loading: true,
    menuAnchorEl: null,
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

  openMenu = (event: { currentTarget: Element }) => this.setState({
    menuAnchorEl: event.currentTarget,
  });

  closeMenu = () => this.setState({
    menuAnchorEl: null,
  });

  manageBilling = () => {
    return request().get(`/billing/portal`)
      .then(result => {
        window.location.assign(result.data.url);
      })
      .catch(error => {
        alert(error);
      });
  };

  doLogout = () => {
    this.props.logout();
    this.props.history.push('/login');
  };

  gotoAccount = () => {
    this.setState({
      menuAnchorEl: null,
    }, () => {
      this.props.history.push('/account');
    });
  };

  renderSubRoutes = () => {
    if (this.props.hasAnyLinks) {
      return this.renderSetup();
    }

    return this.renderNotSetup()
  };

  renderNotSetup = () => {
    return (
      <Switch>
        <Route path="/logout">
          <Logout/>
        </Route>
        <Route path="/setup">
          <InitialPlaidSetup/>
        </Route>
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
        <AppBar position="static">
          <Toolbar>
            <BankAccountSelector/>
            <Button to="/transactions" component={ RouterLink } color="inherit">Transactions</Button>
            <Button to="/expenses" component={ RouterLink } color="inherit">Expenses</Button>
            <Button to="/goals" component={ RouterLink } color="inherit">Goals</Button>
            <BalanceNavDisplay/>
            <div style={ { marginLeft: 'auto' } }/>
            <IconButton onClick={ this.openMenu } edge="start" color="inherit" aria-label="menu">
              <MenuIcon/>
            </IconButton>
            <Menu
              id="user-menu"
              anchorEl={ this.state.menuAnchorEl }
              keepMounted
              open={ Boolean(this.state.menuAnchorEl) }
              onClose={ this.closeMenu }
            >
              <MenuItem
                onClick={ this.manageBilling }
              >
                <CreditCard className="mr-2"/>
                Billing
              </MenuItem>
              <MenuItem
                onClick={ this.doLogout }
              >
                <ExitToApp className="mr-2"/>
                Logout
              </MenuItem>
            </Menu>
          </Toolbar>
        </AppBar>
        <Switch>
          <Route path="/register">
            <Redirect to="/"/>
          </Route>
          <Route path="/login">
            <Redirect to="/"/>
          </Route>
          <Route path="/logout">
            <Logout/>
          </Route>
          <Route path="/transactions">
            <TransactionsView/>
          </Route>
          <Route path="/expenses">
            <ExpensesView/>
          </Route>
          <Route path="/goals">
            <GoalsView/>
          </Route>
          <Route path="/account">
            <AccountView/>
          </Route>
          <Route path="/accounts">
            <AllAccountsView/>
          </Route>
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

    return (
      <Switch>
        <Route path="/account/subscribe/after">
          <AfterCheckout/>
        </Route>
        { this.renderSubRoutes() }
      </Switch>
    );
  }
}

export default connect(
  state => ({
    hasAnyLinks: getHasAnyLinks(state),
  }),
  {
    logout,
    fetchBalances,
    fetchBankAccounts,
    fetchFundingSchedulesIfNeeded,
    fetchSpending,
    fetchLinksIfNeeded,
    fetchInitialTransactionsIfNeeded,
  }
)(withRouter(AuthenticatedApp));
