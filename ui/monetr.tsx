import React, { Fragment } from 'react';
import { Navigate, Outlet, Route, Routes } from 'react-router-dom';

import { useLinksSink } from 'hooks/links';
import { useAppConfigurationSink } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import Loading from 'loading';
import SubscribePage from 'pages/account/subscribe';
import AfterCheckoutPage from 'pages/account/subscribe/after';
import ConfigError from 'pages/error/config';
import LoginNew from 'pages/login-new';
import LogoutPage from 'pages/logout';
import BankSidebar from 'pages/new/BankSidebar';
import BudgetingSidebar from 'pages/new/BudgetingSidebar';
import ExpenseList from 'pages/new/ExpenseList';
import TransactionList from 'pages/new/TransactionList';
import ForgotPasswordNew from 'pages/password/forgot-new';
import RegisterNew from 'pages/register-new';
import SettingsPage from 'pages/settings';
import SetupPage from 'pages/setup';
import OAuthRedirect from 'views/FirstTimeSetup/OAuthRedirect';

export default function Monetr(): JSX.Element {
  const { result: config, isLoading: configIsLoading, isError: configIsError } = useAppConfigurationSink();
  const { isLoading: authIsLoading, result: { user, isActive } } = useAuthenticationSink();
  const { isLoading: linksIsLoading, result: links } = useLinksSink();
  const isAuthenticated = !!user;
  // If the config or authentication is loading just show a loading page.
  if (configIsLoading || authIsLoading || linksIsLoading) {
    return <Loading />;
  }

  // If the config fails to load or is simply not present, show the error page.
  if (configIsError || !config) {
    return <ConfigError />;
  }

  if (!isAuthenticated) {
    return (
      <Routes>
        <Route path='/login' element={ <LoginNew /> } />
        { config?.allowSignUp && <Route path='/register' element={ <RegisterNew /> } /> }
        { config?.allowForgotPassword && <Route path='/password/forgot' element={ <ForgotPasswordNew /> } /> }
        <Route path='/' element={ <Navigate replace to="/login" /> } />
      </Routes>
    );
  }

  if (!isActive) {
    return (
      <Routes>
        <Route path="/logout" element={ <LogoutPage /> } />
        <Route path="/account/subscribe" element={ <SubscribePage /> } />
        <Route path="/account/subscribe/after" element={ <AfterCheckoutPage /> } />
        <Route path="*" element={ <Navigate replace to="/account/subscribe" /> } />
      </Routes>
    );
  }

  const hasAnyLinks = links.size > 0;
  if (!hasAnyLinks) {
    return (
      <Routes>
        <Route path="/logout" element={ <LogoutPage /> } />
        <Route path="/setup" element={ <SetupPage /> } />
        <Route path="/plaid/oauth-return" element={ <OAuthRedirect /> } />
        <Route index path="/" element={ <Navigate replace to="/setup" /> } />
      </Routes>
    );
  }

  return (
    <div className='w-full h-full dark:bg-dark-monetr-background flex'>
      <BankSidebar />
      <div className='w-full h-full flex min-w-0'>
        <Routes>
          <Route element={ <BudgetingLayout /> }>
            <Route path='/transactions' element={ <TransactionList /> } />
            <Route path='/expenses' element={ <ExpenseList /> } />
          </Route>
          <Route path='/settings' element={ <SettingsPage /> } />
          <Route path="/logout" element={ <LogoutPage /> } />
          <Route path="/plaid/oauth-return" element={ <OAuthRedirect /> } />
          <Route path="/setup" element={ <Navigate replace to="/transactions" /> } />
          <Route index path="/" element={ <Navigate replace to="/transactions" /> } />
        </Routes>
      </div>
    </div>
  );
}

function BudgetingLayout(): JSX.Element {
  return (
    <Fragment>
      <BudgetingSidebar />
      <div className='w-full h-full min-w-0 flex flex-col'>
        <Outlet />
      </div>
    </Fragment>
  );
}
