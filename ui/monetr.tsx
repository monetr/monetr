import React, { Fragment } from 'react';
import { Navigate, Outlet, Route, Routes } from 'react-router-dom';

import { useBankAccountsSink } from 'hooks/bankAccounts';
import { useLinksSink } from 'hooks/links';
import { useAppConfigurationSink } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import Loading from 'loading';
import SubscribePage from 'pages/account/subscribe';
import AfterCheckoutPage from 'pages/account/subscribe/after';
import ConfigError from 'pages/error/config';
import ExpenseDetails from 'pages/expense/details';
import FundingDetails from 'pages/funding/details';
import FundingNew from 'pages/funding-new';
import GoalsNew from 'pages/goals-new';
import LoginNew from 'pages/login-new';
import LogoutPage from 'pages/logout';
import BankSidebar from 'pages/new/BankSidebar';
import BudgetingSidebar from 'pages/new/BudgetingSidebar';
import ExpenseList from 'pages/new/ExpenseList';
import MobileSidebar from 'pages/new/MobileSidebar';
import TransactionList from 'pages/new/TransactionList';
import ForgotPasswordNew from 'pages/password/forgot-new';
import PasswordResetNew from 'pages/password/reset-new';
import RegisterNew from 'pages/register-new';
import SettingsPage from 'pages/settings';
import SetupPage from 'pages/setup';
import TransactionDetails from 'pages/transaction/details';
import VerifyEmail from 'pages/verify/email';
import ResendVerificationPage from 'pages/verify/email/resend';
import OAuthRedirect from 'views/FirstTimeSetup/OAuthRedirect';

export default function Monetr(): JSX.Element {
  const {
    result: config,
    isLoading: configIsLoading,
    isError: configIsError,
  } = useAppConfigurationSink();
  const { isLoading: authIsLoading, result: { user, isActive } } = useAuthenticationSink();
  const { isLoading: linksIsLoading, isFetching: linksIsFetching, data: links } = useLinksSink();
  const isAuthenticated = !!user;
  // If the config or authentication is loading just show a loading page.
  if (configIsLoading || authIsLoading || (linksIsLoading && linksIsFetching)) {
    return <Loading />;
  }

  // If the config fails to load or is simply not present, show the error page.
  if (configIsError) {
    return <ConfigError />;
  }

  if (!isAuthenticated) {
    return (
      <Routes>
        <Route path='/login' element={ <LoginNew /> } />
        <Route path="/logout" element={ <LogoutPage /> } />
        {config?.allowSignUp && <Route path='/register' element={ <RegisterNew /> } />}
        {config?.allowForgotPassword && <Route path='/password/forgot' element={ <ForgotPasswordNew /> } />}
        <Route path="/password/reset" element={ <PasswordResetNew /> } />
        <Route path="/verify/email" element={ <VerifyEmail /> } />
        <Route path="/verify/email/resend" element={ <ResendVerificationPage /> } />
        <Route path='/' element={ <Navigate replace to="/login" /> } />
        <Route path='*' element={ <Navigate replace to="/login" /> } />
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

  const hasAnyLinks = links.length > 0;
  if (!hasAnyLinks) {
    return (
      <Routes>
        <Route path="/logout" element={ <LogoutPage /> } />
        <Route path="/setup" element={ <SetupPage /> } />
        <Route path="/plaid/oauth-return" element={ <OAuthRedirect /> } />
        <Route path="/account/subscribe/after" element={ <Navigate replace to="/setup" /> } />
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
          <Route path='/bank/:bankAccountId' element={ <BudgetingLayout /> }>
            <Route path='transactions' element={ <TransactionList /> } />
            <Route path='transactions/:transactionId/details' element={ <TransactionDetails /> } />
            <Route path='expenses' element={ <ExpenseList /> } />
            <Route path='expenses/:spendingId/details' element={ <ExpenseDetails /> } />
            <Route path='goals' element={ <GoalsNew /> } />
            <Route path='funding' element={ <FundingNew /> } />
            <Route path='funding/:fundingId/details' element={ <FundingDetails /> } />
          </Route>
          <Route path='/settings' element={ <SettingsPage /> } />
          <Route path="/logout" element={ <LogoutPage /> } />
          <Route path="/plaid/oauth-return" element={ <OAuthRedirect /> } />
          <Route path="/setup" element={ <Navigate replace to="/" /> } />
          <Route index path="/" element={ <RedirectToBank /> } />
        </Routes>
      </div>
    </div>
  );
}

function BudgetingLayout(): JSX.Element {
  return (
    <Fragment>
      <BudgetingSidebar className='hidden lg:flex' />
      <div className='w-full h-full min-w-0 flex flex-col'>
        <Outlet />
      </div>
    </Fragment>
  );
}

function RedirectToBank(): JSX.Element {
  const { data: links, isLoading: linksIsLoading } = useLinksSink();
  const { result: bankAccounts, isLoading: bankAccountsIsLoading } = useBankAccountsSink();
  if (linksIsLoading || bankAccountsIsLoading) {
    return null;
  }
  if (links.length === 0) {
    return null;
  }

  const link = links[0];
  const accounts = Array.from(bankAccounts.values())
    .filter(account => account.linkId === link.linkId)
    .sort((a, b) => {
      const items = [a, b];
      const values = [
        0, // a
        0, // b
      ];
      for (let i = 0; i < 2; i++) {
        const item = items[i];
        if (item.accountType === 'depository') {
          values[i] += 2;
        }
        switch (item.accountSubType) {
          case 'checking':
            values[i] += 2;
            break;
          case 'savings':
            values[i] += 1;
            break;
        }
      }

      return values[0];
    });

  if (accounts.length === 0) {
    return null;
  }

  const account = accounts[0];

  return <Navigate replace to={ `/bank/${account.bankAccountId}/transactions` } />;
}
