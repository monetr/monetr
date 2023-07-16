import React, { Fragment } from 'react';
import { Navigate, Outlet, Route, Routes } from 'react-router-dom';

import { useLinksSink } from 'hooks/links';
import { useAppConfigurationSink } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import Loading from 'loading';
import SubscribePage from 'pages/account/subscribe';
import AfterCheckoutPage from 'pages/account/subscribe/after';
import ConfigError from 'pages/error/config';
import ExpenseDetails from 'pages/expense/details';
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
import TransactionDetails from 'pages/transaction/details';
import OAuthRedirect from 'views/FirstTimeSetup/OAuthRedirect';
import MobileSidebar from 'pages/new/MobileSidebar';

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
        <Route path="/logout" element={ <LogoutPage /> } />
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

  // TODO Fix banksidebar issue by moving it into the godless react router routes
  return (
    <div className='max-w-screen max-h-screen h-full w-full dark:bg-dark-monetr-background flex'>
      <BankSidebar className='hidden lg:flex' />
      <MobileSidebar />
      <div className='w-full h-full flex min-w-0'>
        <Routes>
          <Route path='/bank/:bankAccountId' element={ <BudgetingLayout className='hidden lg:flex' /> }>
            <Route path='transactions' element={ <TransactionList /> } />
            <Route path='transactions/:transactionId/details' element={ <TransactionDetails /> } />
            <Route path='expenses' element={ <ExpenseList /> } />
            <Route path='expenses/:spendingId/details' element={ <ExpenseDetails /> } />
          </Route>
          <Route path='/settings' element={ <SettingsPage /> } />
          <Route path="/logout" element={ <LogoutPage /> } />
          <Route path="/plaid/oauth-return" element={ <OAuthRedirect /> } />
          <Route path="/setup" element={ <Navigate replace to="/" /> } />
          <Route index path="/" element={ <Navigate replace to="/transactions" /> } />
        </Routes>
      </div>
    </div>
  );
}

interface BudgetingLayoutProps {
  className?: string;
}

function BudgetingLayout(props: BudgetingLayoutProps): JSX.Element {
  return (
    <Fragment>
      <BudgetingSidebar className={ props.className } />
      <div className='w-full h-full min-w-0 flex flex-col'>
        <Outlet />
      </div>
    </Fragment>
  );
}
