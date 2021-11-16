import NavigationBar from 'NavigationBar';
import React, { Fragment, useEffect, useState } from 'react';
import { getHasAnyLinks } from 'shared/links/selectors/getHasAnyLinks';
import fetchBalances from 'shared/balances/actions/fetchBalances';
import fetchBankAccounts from 'shared/bankAccounts/actions/fetchBankAccounts';
import { fetchFundingSchedulesIfNeeded } from 'shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded';
import fetchSpending from 'shared/spending/actions/fetchSpending';
import fetchLinksIfNeeded from 'shared/links/actions/fetchLinksIfNeeded';
import fetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import { Navigate, Route, Routes } from 'react-router-dom';
import { useSelector, useStore } from 'react-redux';
import { Backdrop, CircularProgress } from '@mui/material';
import TransactionsView from 'views/Transactions/TransactionsView';
import ExpensesView from 'views/Expenses/ExpensesView';
import GoalsView from 'views/Goals/GoalsView';
import OAuthRedirect from 'views/FirstTimeSetup/OAuthRedirect';
import AllAccountsView from 'views/AccountView/AllAccountsView';
import Logout from 'views/Authentication/Logout';
import InitialPlaidSetup from 'views/Setup/InitialPlaidSetup';

const AuthenticatedApp = (): JSX.Element => {
  const [loading, setLoading] = useState(true);
  const { dispatch, getState } = useStore();

  useEffect(() => {
    Promise.all([
      fetchLinksIfNeeded()(dispatch, getState),
      fetchBankAccounts()(dispatch).then(() => Promise.all([
        fetchInitialTransactionsIfNeeded()(dispatch, getState),
        fetchFundingSchedulesIfNeeded()(dispatch, getState),
        fetchSpending()(dispatch, getState),
        fetchBalances()(dispatch, getState),
      ])),
    ])
      .finally(() => setLoading(false));
  }, []); // No dependencies makes sure that this will only ever be called the first time this component is mounted.

  const hasAnyLinks = useSelector(getHasAnyLinks);

  if (loading) {
    return (
      <Backdrop open={ true }>
        <CircularProgress color="inherit"/>
      </Backdrop>
    );
  }

  // If the user has no links setup then we want to only give them a handful of routes to get things setup.
  if (!hasAnyLinks) {
    return (
      <Routes>
        <Route path="/logout" element={ <Logout/> }/>
        <Route path="/setup" element={ <InitialPlaidSetup/> }/>
        <Route path="/plaid/oauth-return" element={ <OAuthRedirect/> }/>
        <Route path="*" element={ <Navigate replace to="/setup"/> }/>
      </Routes>
    );
  }

  return (
    <Fragment>
      <NavigationBar/>
      <Routes>
        <Route path="/register" element={ <Navigate replace to="/"/> }/>
        <Route path="/login" element={ <Navigate replace to="/"/> }/>
        <Route path="/logout" element={ <Logout/> }/>
        <Route path="/transactions" element={ <TransactionsView/> }/>
        <Route path="/expenses" element={ <ExpensesView/> }/>
        <Route path="/goals" element={ <GoalsView/> }/>
        <Route path="/accounts" element={ <AllAccountsView/> }/>
        <Route path="*" element={ <Navigate replace to="/transactions"/> }/>
      </Routes>
    </Fragment>
  );
};

export default AuthenticatedApp;