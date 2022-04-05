import Sidebar from 'components/Layout/Sidebar/Sidebar';
import NavigationBar from 'components/Layout/NavigationBar/NavigationBar';
import AccountsPage from 'pages/accounts';
import ExpensesPage from 'pages/expenses';
import GoalsPage from 'pages/goals';
import LogoutPage from 'pages/logout';
import SettingsPage from 'pages/settings';
import SubscriptionPage from 'pages/subscription';
import React, { Fragment, useState } from 'react';
import useFetchLinksIfNeeded from 'shared/links/hooks/useFetchLinksIfNeeded';
import { getHasAnyLinks } from 'shared/links/selectors/getHasAnyLinks';
import fetchBalances from 'shared/balances/actions/fetchBalances';
import fetchBankAccounts from 'shared/bankAccounts/actions/fetchBankAccounts';
import { fetchFundingSchedulesIfNeeded } from 'shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded';
import fetchSpending from 'shared/spending/actions/fetchSpending';
import useFetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import { Navigate, Route, Routes } from 'react-router-dom';
import { useSelector, useStore } from 'react-redux';
import { Backdrop, CircularProgress } from '@mui/material';
import useMountEffect from 'shared/util/useMountEffect';
import TransactionsView from 'views/Transactions/TransactionsView';
import GoalsView from 'components/Goals/GoalsView/GoalsView';
import OAuthRedirect from 'views/FirstTimeSetup/OAuthRedirect';
import InitialPlaidSetup from 'views/Setup/InitialPlaidSetup';

const AuthenticatedApp = (): JSX.Element => {
  const [loading, setLoading] = useState(true);
  const { dispatch, getState } = useStore();

  const fetchInitialTransactionsIfNeeded = useFetchInitialTransactionsIfNeeded();
  const fetchLinksIfNeeded = useFetchLinksIfNeeded();

  useMountEffect(() => {
    Promise.all([
      fetchLinksIfNeeded(),
      fetchBankAccounts()(dispatch).then(() => Promise.all([
        fetchInitialTransactionsIfNeeded(),
        fetchFundingSchedulesIfNeeded()(dispatch, getState),
        fetchSpending()(dispatch, getState),
        fetchBalances()(dispatch, getState),
      ])),
    ])
      .finally(() => setLoading(false));
  });

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
        <Route path="/logout" element={ <LogoutPage/> }/>
        <Route path="/setup" element={ <InitialPlaidSetup/> }/>
        <Route path="/plaid/oauth-return" element={ <OAuthRedirect/> }/>
        <Route path="*" element={ <Navigate replace to="/setup"/> }/>
      </Routes>
    );
  }

  return (
    <Fragment>
      <div className="flex h-full min-w-0 min-h-full">
        <Sidebar/>
        <div className="relative flex flex-col flex-1 w-0 min-w-0 mb-8 lg:ml-64">
          <NavigationBar/>
          <Routes>
            <Route path="/register" element={ <Navigate replace to="/"/> }/>
            <Route path="/login" element={ <Navigate replace to="/"/> }/>
            <Route path="/logout" element={ <LogoutPage/> }/>
            <Route path="/transactions" element={ <TransactionsView/> }/>
            <Route path="/expenses" element={ <ExpensesPage/> }/>
            <Route path="/goals" element={ <GoalsPage/> }/>
            <Route path="/accounts" element={ <AccountsPage/> }/>
            <Route path="/settings" element={ <SettingsPage/> }/>
            <Route path="/subscription" element={ <SubscriptionPage/> }/>
            <Route path="*" element={ <Navigate replace to="/transactions"/> }/>
          </Routes>
        </div>
      </div>
    </Fragment>
  );
};

export default AuthenticatedApp;
