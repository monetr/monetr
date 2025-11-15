import { Fragment } from 'react';
import { withSentryReactRouterV6Routing } from '@sentry/react';
import { Navigate, Outlet, Route, Routes } from 'react-router-dom';

import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';
import MobileSidebarContextProvider from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import SettingsLayout from '@monetr/interface/components/Layout/SettingsLayout';
import Sidebar from '@monetr/interface/components/Layout/Sidebar';
import PlaidSetup from '@monetr/interface/components/setup/PlaidSetup';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { useBankAccounts } from '@monetr/interface/hooks/useBankAccounts';
import { useLinks } from '@monetr/interface/hooks/useLinks';
import Loading from '@monetr/interface/loading';
import SubscribePage from '@monetr/interface/pages/account/subscribe';
import AfterCheckoutPage from '@monetr/interface/pages/account/subscribe/after';
import BankAccountSettingsPage from '@monetr/interface/pages/bank/settings';
import ConfigError from '@monetr/interface/pages/error/config';
import ExpenseDetails from '@monetr/interface/pages/expense/details';
import Expenses from '@monetr/interface/pages/expenses';
import Funding from '@monetr/interface/pages/funding';
import FundingDetails from '@monetr/interface/pages/funding/details';
import Goals from '@monetr/interface/pages/goals';
import GoalDetails from '@monetr/interface/pages/goals/details';
import LinkCreatePage from '@monetr/interface/pages/link/create';
import CreateManualLinkPage from '@monetr/interface/pages/link/create/manual';
import LinkDetails from '@monetr/interface/pages/link/details';
import Login from '@monetr/interface/pages/login';
import MultifactorAuthenticationPage from '@monetr/interface/pages/login/multifactor';
import LogoutPage from '@monetr/interface/pages/logout';
import ForgotPassword from '@monetr/interface/pages/password/forgot';
import PasswordReset from '@monetr/interface/pages/password/reset';
import OauthReturn from '@monetr/interface/pages/plaid/oauth-return';
import Register from '@monetr/interface/pages/register';
import SettingsAbout from '@monetr/interface/pages/settings/about';
import SettingsBilling from '@monetr/interface/pages/settings/billing';
import SettingsOverview from '@monetr/interface/pages/settings/overview';
import SettingsSecurity from '@monetr/interface/pages/settings/security';
import SetupPage from '@monetr/interface/pages/setup';
import SetupManualLinkPage from '@monetr/interface/pages/setup/manual';
import SubscriptionPage from '@monetr/interface/pages/subscription';
import TransactionDetails from '@monetr/interface/pages/transaction/details';
import Transactions from '@monetr/interface/pages/transactions';
import VerifyEmail from '@monetr/interface/pages/verify/email';
import ResendVerificationPage from '@monetr/interface/pages/verify/email/resend';
import sortAccounts from '@monetr/interface/util/sortAccounts';
import BudgetingLayout from '@monetr/interface/components/Layout/BudgetLayout';

const RoutesImpl = withSentryReactRouterV6Routing(Routes);

export default function Monetr(): JSX.Element {
  const { data: config, isLoading: configIsLoading, isError: configIsError } = useAppConfiguration();
  const { isLoading: authIsLoading, data: auth } = useAuthentication();
  const { isLoading: linksIsLoading, data: links } = useLinks();
  const { isLoading: bankAccountsIsLoading } = useBankAccounts();

  const isAuthenticated = Boolean(auth?.user);
  // If the config or authentication is loading just show a loading page.
  // Links is loading is weird becuase the loading state will be true until we actually request links. But links won't
  // be requested until we are authenticated with an active subscription.
  if (
    configIsLoading ||
    authIsLoading ||
    ((bankAccountsIsLoading || linksIsLoading) && auth?.isActive && !auth?.mfaPending)
  ) {
    return <Loading />;
  }

  // If the config fails to load or is simply not present, show the error page.
  if (configIsError) {
    return <ConfigError />;
  }

  if (!isAuthenticated) {
    return (
      <RoutesImpl>
        <Route element={<Login />} path='/login' />
        <Route element={<LogoutPage />} path='/logout' />
        {config?.allowSignUp && <Route element={<Register />} path='/register' />}
        {config?.allowForgotPassword && <Route element={<ForgotPassword />} path='/password/forgot' />}
        <Route element={<PasswordReset />} path='/password/reset' />
        <Route element={<VerifyEmail />} path='/verify/email' />
        <Route element={<ResendVerificationPage />} path='/verify/email/resend' />
        <Route element={<Navigate replace to='/login' />} path='/' />
        <Route element={<Navigate replace to='/login' />} path='*' />
      </RoutesImpl>
    );
  }

  // If the currently authenticated user requires MFA then only allow them to access the MFA pages.
  if (auth?.mfaPending) {
    return (
      <RoutesImpl>
        <Route element={<MultifactorAuthenticationPage />} path='/login/multifactor' />
        <Route element={<LogoutPage />} path='/logout' />
        <Route element={<Navigate replace to='/login/multifactor' />} path='*' />
      </RoutesImpl>
    );
  }

  if (!auth?.isActive) {
    return (
      <RoutesImpl>
        <Route element={<LogoutPage />} path='/logout' />
        <Route element={<SubscribePage />} path='/account/subscribe' />
        <Route element={<AfterCheckoutPage />} path='/account/subscribe/after' />
        <Route element={<Navigate replace to='/account/subscribe' />} path='*' />
      </RoutesImpl>
    );
  }

  const hasAnyLinks = links?.length > 0;
  if (!hasAnyLinks) {
    return (
      <RoutesImpl>
        <Route element={<LogoutPage />} path='/logout' />
        <Route element={<SetupPage manualEnabled={config?.manualEnabled} />} path='/setup' />
        <Route element={<PlaidSetup alreadyOnboarded />} path='/setup/plaid' />
        <Route element={<SetupManualLinkPage />} path='/setup/manual' />
        <Route element={<OauthReturn />} path='/plaid/oauth-return' />
        <Route element={<Navigate replace to='/setup' />} path='/account/subscribe/after' />
        <Route element={<Navigate replace to='/setup' />} index path='*' />
      </RoutesImpl>
    );
  }

  return (
    <MobileSidebarContextProvider>
      <Sidebar />
      <RoutesImpl>
        <Route element={<BudgetingLayout />} path='/bank/:bankAccountId'>
          <Route element={<BankAccountSettingsPage />} path='settings' />
          <Route element={<Transactions />} path='transactions' />
          <Route element={<TransactionDetails />} path='transactions/:transactionId/details' />
          <Route element={<Expenses />} path='expenses' />
          <Route element={<ExpenseDetails />} path='expenses/:spendingId/details' />
          <Route element={<Goals />} path='goals' />
          <Route element={<GoalDetails />} path='goals/:spendingId/details' />
          <Route element={<Funding />} path='funding' />
          <Route element={<FundingDetails />} path='funding/:fundingId/details' />
        </Route>
        <Route element={<SettingsLayout />} path='/settings'>
          <Route element={<Navigate replace to='/settings/overview' />} path='' />
          <Route element={<SettingsOverview />} path='overview' />
          <Route element={<SettingsSecurity />} path='security' />
          {config?.billingEnabled && <Route element={<SettingsBilling />} path='billing' />}
          <Route element={<SettingsAbout />} path='about' />
        </Route>
        <Route element={<LinkDetails />} path='/link/:linkId/details' />
        <Route element={<LinkCreatePage />} path='/link/create' />
        <Route element={<PlaidSetup alreadyOnboarded />} path='/link/create/plaid' />
        <Route element={<CreateManualLinkPage />} path='/link/create/manual' />
        <Route element={<LogoutPage />} path='/logout' />
        <Route element={<OauthReturn />} path='/plaid/oauth-return' />
        <Route element={<SubscriptionPage />} path='/subscription' />
        <Route element={<Navigate replace to='/' />} path='/account/subscribe' />
        <Route element={<AfterCheckoutPage />} path='/account/subscribe/after' />
        <Route element={<Navigate replace to='/' />} path='/setup' />
        <Route element={<Navigate replace to='/' />} path='/password/reset' />
        <Route element={<Navigate replace to='/' />} path='/register' />
        <Route element={<Navigate replace to='/' />} path='/login' />
        <Route element={<Navigate replace to='/' />} path='/login/multifactor' />
        <Route element={<RedirectToBank />} index path='/' />
      </RoutesImpl>
    </MobileSidebarContextProvider>
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

  const linksSorted = links.sort((a, b) => {
    const nameA = a.getName().toUpperCase();
    const nameB = b.getName().toUpperCase();
    if (nameA < nameB) {
      return -1;
    }
    if (nameA > nameB) {
      return 1;
    }

    // names must be equal
    return 0;
  });

  const link = linksSorted[0];
  const accounts = sortAccounts(Array.from(bankAccounts.values()).filter(account => account.linkId === link.linkId));

  if (accounts.length === 0) {
    return <Navigate replace to='/link/create' />;
  }

  const account = accounts[0];

  return <Navigate replace to={`/bank/${account.bankAccountId}/transactions`} />;
}
