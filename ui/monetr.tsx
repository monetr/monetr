import React, { Fragment } from 'react';
import { Navigate, Outlet, Route, Routes } from 'react-router-dom';

import BankSidebar from 'components/Layout/BankSidebar';
import BudgetingSidebar from 'components/Layout/BudgetingSidebar';
import MobileSidebar from 'components/Layout/MobileSidebar';
import SettingsLayout from 'components/Layout/SettingsLayout';
import { useBankAccounts } from 'hooks/bankAccounts';
import { useLinks } from 'hooks/links';
import { useAppConfigurationSink } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import Loading from 'loading';
import SubscribePage from 'pages/account/subscribe';
import AfterCheckoutPage from 'pages/account/subscribe/after';
import ConfigError from 'pages/error/config';
import ExpenseDetails from 'pages/expense/details';
import Expenses from 'pages/expenses';
import Funding from 'pages/funding';
import FundingDetails from 'pages/funding/details';
import Goals from 'pages/goals';
import GoalDetails from 'pages/goals/details';
import LinkCreatePage from 'pages/link/create';
import Login from 'pages/login';
import LogoutPage from 'pages/logout';
import ForgotPasswordNew from 'pages/password/forgot';
import PasswordResetNew from 'pages/password/reset';
import OauthReturn from 'pages/plaid/oauth-return';
import Register from 'pages/register';
import SettingsAbout from 'pages/settings/about';
import SettingsBilling from 'pages/settings/billing';
import SettingsOverview from 'pages/settings/overview';
import SettingsSecurity from 'pages/settings/security';
import SetupPage from 'pages/setup';
import SubscriptionPage from 'pages/subscription';
import TransactionDetails from 'pages/transaction/details';
import Transactions from 'pages/transactions';
import VerifyEmail from 'pages/verify/email';
import ResendVerificationPage from 'pages/verify/email/resend';
import sortAccounts from 'util/sortAccounts';

export default function Monetr(): JSX.Element {
  const {
    result: config,
    isLoading: configIsLoading,
    isError: configIsError,
  } = useAppConfigurationSink();
  const { isLoading: authIsLoading, result: { user, isActive } } = useAuthenticationSink();
  const { isLoading: linksIsLoading, data: links } = useLinks();
  const isAuthenticated = !!user;
  // If the config or authentication is loading just show a loading page.
  // Links is loading is weird becuase the loading state will be true until we actually request links. But links won't
  // be requested until we are authenticated with an active subscription.
  if (configIsLoading || authIsLoading || (linksIsLoading && isActive)) {
    return <Loading />;
  }

  // If the config fails to load or is simply not present, show the error page.
  if (configIsError) {
    return <ConfigError />;
  }

  if (!isAuthenticated) {
    return (
      <Routes>
        <Route path='/login' element={ <Login /> } />
        <Route path="/logout" element={ <LogoutPage /> } />
        {config?.allowSignUp && <Route path='/register' element={ <Register /> } />}
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
        <Route path="/plaid/oauth-return" element={ <OauthReturn /> } />
        <Route path="/account/subscribe/after" element={ <Navigate replace to="/setup" /> } />
        <Route index path="/" element={ <Navigate replace to="/setup" /> } />
      </Routes>
    );
  }

  return (
    <div className='max-w-screen max-h-screen h-full w-full dark:bg-dark-monetr-background flex'>
      <BankSidebar className='hidden lg:flex' />
      <MobileSidebar />
      <div className='w-full h-full flex min-w-0 overflow-y-auto'>
        <Routes>
          <Route path='/bank/:bankAccountId' element={ <BudgetingLayout /> }>
            <Route path='transactions' element={ <Transactions /> } />
            <Route path='transactions/:transactionId/details' element={ <TransactionDetails /> } />
            <Route path='expenses' element={ <Expenses /> } />
            <Route path='expenses/:spendingId/details' element={ <ExpenseDetails /> } />
            <Route path='goals' element={ <Goals /> } />
            <Route path='goals/:spendingId/details' element={ <GoalDetails /> } />
            <Route path='funding' element={ <Funding /> } />
            <Route path='funding/:fundingId/details' element={ <FundingDetails /> } />
          </Route>
          <Route path='/settings' element={ <SettingsLayout /> }>
            <Route path='' element={ <Navigate replace to="/settings/overview" /> } />
            <Route path='overview' element={ <SettingsOverview /> } />
            <Route path='security' element={ <SettingsSecurity /> } />
            { config?.billingEnabled && (
              <Route path='billing' element={ <SettingsBilling /> } />
            ) }
            <Route path='about' element={ <SettingsAbout /> } />
          </Route>
          <Route path='/link/create' element={ <LinkCreatePage /> } />
          <Route path="/logout" element={ <LogoutPage /> } />
          <Route path="/plaid/oauth-return" element={ <OauthReturn /> } />
          <Route path='/subscription' element={ <SubscriptionPage /> } />
          <Route path='/account/subscribe/after' element={ <Navigate replace to="/" /> } />
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
  const { data: links, isLoading: linksIsLoading } = useLinks();
  const { data: bankAccounts, isLoading: bankAccountsIsLoading } = useBankAccounts();
  if (linksIsLoading || bankAccountsIsLoading) {
    return null;
  }
  if (links.length === 0) {
    return null;
  }

  const link = links[0];
  const accounts = sortAccounts(Array.from(bankAccounts.values()).filter(account => account.linkId === link.linkId));

  if (accounts.length === 0) {
    return null;
  }

  const account = accounts[0];

  return <Navigate replace to={ `/bank/${account.bankAccountId}/transactions` } />;
}
