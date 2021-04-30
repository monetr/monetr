import {
  AppBar,
  Backdrop,
  Button,
  CircularProgress,
  IconButton,
  Menu,
  MenuItem,
  Toolbar,
  Typography
} from "@material-ui/core";
import MenuIcon from "@material-ui/icons/Menu";
import BalanceNavDisplay from "components/Balance/BalanceNavDisplay";
import BankAccountSelector from "components/BankAccountSelector";
import PropTypes from "prop-types";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import { Link as RouterLink, Redirect, Route, Switch, withRouter } from "react-router-dom";
import { bindActionCreators } from "redux";
import logout from "shared/authentication/actions/logout";
import fetchBalances from "shared/balances/actions/fetchBalances";
import fetchBankAccounts from "shared/bankAccounts/actions/fetchBankAccounts";
import { fetchFundingSchedulesIfNeeded } from "shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded";
import fetchLinksIfNeeded from "shared/links/actions/fetchLinksIfNeeded";
import { getHasAnyLinks } from "shared/links/selectors/getHasAnyLinks";
import fetchSpending from "shared/spending/actions/fetchSpending";
import fetchInitialTransactionsIfNeeded from "shared/transactions/actions/fetchInitialTransactionsIfNeeded";
import ExpensesView from "views/ExpensesView";
import FirstTimeSetup from "views/FirstTimeSetup";
import GoalsView from "views/GoalsView";
import TransactionsView from "views/TransactionsView";
import AccountView from "views/AccountView";

export class AuthenticatedApplication extends Component {
  state = {
    loading: true,
    anchorEl: null,
  };

  static propTypes = {
    logout: PropTypes.func.isRequired,
    history: PropTypes.object.isRequired,
    location: PropTypes.instanceOf(Location).isRequired,
    fetchLinksIfNeeded: PropTypes.func.isRequired,
    fetchBankAccounts: PropTypes.func.isRequired,
    fetchSpending: PropTypes.func.isRequired,
    fetchFundingSchedulesIfNeeded: PropTypes.func.isRequired,
    hasAnyLinks: PropTypes.bool.isRequired,
    fetchInitialTransactionsIfNeeded: PropTypes.func.isRequired,
    fetchBalances: PropTypes.func.isRequired,
  };

  componentDidMount() {
    Promise.all([
      this.props.fetchLinksIfNeeded(),
      this.props.fetchBankAccounts().then(() => {
        return Promise.all([
          this.props.fetchInitialTransactionsIfNeeded(),
          this.props.fetchFundingSchedulesIfNeeded(),
          this.props.fetchSpending(),
          this.props.fetchBalances(),
        ]);
      }),
    ])
      .then(() => this.setState({ loading: false }));
  }

  openMenu = event => {
    this.setState({
      anchorEl: event.currentTarget,
    });
  };

  closeMenu = () => {
    this.setState({
      anchorEl: null,
    });
  };

  doLogout = () => {
    this.props.logout();
    this.props.history.push('/login');
  };

  renderNotSetup = () => {
    return (
      <Switch>
        <Route path="/setup">
          <FirstTimeSetup/>
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

  gotoAccount = () => {
    this.setState({
      anchorEl: null,
    }, () => {
      this.props.history.push('/account');
    });
  };

  renderSetup = () => {
    return (
      <Fragment>
        <AppBar position="static">
          <Toolbar>
            <Button to="/transactions" component={ RouterLink } color="inherit">Transactions</Button>
            <Button to="/expenses" component={ RouterLink } color="inherit">Expenses</Button>
            <Button to="/goals" component={ RouterLink } color="inherit">Goals</Button>
            <BalanceNavDisplay/>
            <div style={ { marginLeft: 'auto' } }/>
            <div style={ { marginRight: '10px', marginLeft: '10px' } }>
              <BankAccountSelector/>
            </div>
            <IconButton onClick={ this.openMenu } edge="start" color="inherit" aria-label="menu">
              <MenuIcon/>
            </IconButton>
            <Menu
              id="user-menu"
              anchorEl={ this.state.anchorEl }
              keepMounted
              open={ Boolean(this.state.anchorEl) }
              onClose={ this.closeMenu }
            >
              <MenuItem onClick={ this.gotoAccount }>My account</MenuItem>
              <MenuItem onClick={ this.doLogout }>Logout</MenuItem>
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
        <Backdrop  open={ true }>
          <CircularProgress color="inherit" />
        </Backdrop>
      );
    }

    if (this.props.hasAnyLinks) {
      return this.renderSetup();
    }

    return this.renderNotSetup()
  }
}

export default connect(
  state => ({
    hasAnyLinks: getHasAnyLinks(state),
  }),
  dispatch => bindActionCreators({
    logout,
    fetchLinksIfNeeded,
    fetchBankAccounts,
    fetchSpending,
    fetchFundingSchedulesIfNeeded,
    fetchInitialTransactionsIfNeeded,
    fetchBalances,
  }, dispatch),
)(withRouter(AuthenticatedApplication));
