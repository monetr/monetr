import { Fragment, lazy } from 'react';
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
import ForgotPasswordNew from '@monetr/interface/pages/password/forgot';
import PasswordResetNew from '@monetr/interface/pages/password/reset';
import Register from '@monetr/interface/pages/register';
import SettingsAbout from '@monetr/interface/pages/settings/about';
import SettingsOverview from '@monetr/interface/pages/settings/overview';
import SettingsSecurity from '@monetr/interface/pages/settings/security';
import SetupPage from '@monetr/interface/pages/setup';
import SetupManualLinkPage from '@monetr/interface/pages/setup/manual';
import TransactionDetails from '@monetr/interface/pages/transaction/details';
import Transactions from '@monetr/interface/pages/transactions';
import VerifyEmail from '@monetr/interface/pages/verify/email';
import ResendVerificationPage from '@monetr/interface/pages/verify/email/resend';
import sortAccounts from '@monetr/interface/util/sortAccounts';

// Billing related pages do not need to be in the main bundle, this way self hosted deployments will in general just run
// better since they never need to load billing related modules.
const AfterCheckoutPage = lazy(() => import('@monetr/interface/pages/account/subscribe/after'));
const SubscribePage = lazy(() => import('@monetr/interface/pages/account/subscribe'));
const SettingsBilling = lazy(() => import('@monetr/interface/pages/settings/billing'));
const SubscriptionPage = lazy(() => import('@monetr/interface/pages/subscription'));

// Plaid OAuth return is barely ever used anymore so it can be broken apart as well.
const OauthReturn = lazy(() => import('@monetr/interface/pages/plaid/oauth-return'));

import styles from './monetr.module.css';

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
        <Route path='/login' element={<Login />} />
        <Route path='/logout' element={<LogoutPage />} />
        {config?.allowSignUp && <Route path='/register' element={<Register />} />}
        {config?.allowForgotPassword && <Route path='/password/forgot' element={<ForgotPasswordNew />} />}
        <Route path='/password/reset' element={<PasswordResetNew />} />
        <Route path='/verify/email' element={<VerifyEmail />} />
        <Route path='/verify/email/resend' element={<ResendVerificationPage />} />
        <Route path='/' element={<Navigate replace to='/login' />} />
        <Route path='*' element={<Navigate replace to='/login' />} />
      </RoutesImpl>
    );
  }

  // If the currently authenticated user requires MFA then only allow them to access the MFA pages.
  if (auth?.mfaPending) {
    return (
      <RoutesImpl>
        <Route path='/login/multifactor' element={<MultifactorAuthenticationPage />} />
        <Route path='/logout' element={<LogoutPage />} />
        <Route path='*' element={<Navigate replace to='/login/multifactor' />} />
      </RoutesImpl>
    );
  }

  if (!auth?.isActive) {
    return (
      <RoutesImpl>
        <Route path='/logout' element={<LogoutPage />} />
        <Route path='/account/subscribe' element={<SubscribePage />} />
        <Route path='/account/subscribe/after' element={<AfterCheckoutPage />} />
        <Route path='*' element={<Navigate replace to='/account/subscribe' />} />
      </RoutesImpl>
    );
  }

  const hasAnyLinks = links?.length > 0;
  if (!hasAnyLinks) {
    return (
      <RoutesImpl>
        <Route path='/logout' element={<LogoutPage />} />
        <Route path='/setup' element={<SetupPage manualEnabled={config?.manualEnabled} />} />
        <Route path='/setup/plaid' element={<PlaidSetup alreadyOnboarded />} />
        <Route path='/setup/manual' element={<SetupManualLinkPage />} />
        <Route path='/plaid/oauth-return' element={<OauthReturn />} />
        <Route path='/account/subscribe/after' element={<Navigate replace to='/setup' />} />
        <Route index path='*' element={<Navigate replace to='/setup' />} />
      </RoutesImpl>
    );
  }

  return (
    <MobileSidebarContextProvider>
      <div className={styles.layout}>
        <Sidebar />
        <div className='w-full h-full flex min-w-0 overflow-y-auto'>
          <RoutesImpl>
            <Route path='/bank/:bankAccountId' element={<BudgetingLayout />}>
              <Route path='settings' element={<BankAccountSettingsPage />} />
              <Route path='transactions' element={<Transactions />} />
              <Route path='transactions/:transactionId/details' element={<TransactionDetails />} />
              <Route path='expenses' element={<Expenses />} />
              <Route path='expenses/:spendingId/details' element={<ExpenseDetails />} />
              <Route path='goals' element={<Goals />} />
              <Route path='goals/:spendingId/details' element={<GoalDetails />} />
              <Route path='funding' element={<Funding />} />
              <Route path='funding/:fundingId/details' element={<FundingDetails />} />
            </Route>
            <Route path='/settings' element={<SettingsLayout />}>
              <Route path='' element={<Navigate replace to='/settings/overview' />} />
              <Route path='overview' element={<SettingsOverview />} />
              <Route path='security' element={<SettingsSecurity />} />
              {config?.billingEnabled && <Route path='billing' element={<SettingsBilling />} />}
              <Route path='about' element={<SettingsAbout />} />
            </Route>
            <Route path='/link/:linkId/details' element={<LinkDetails />} />
            <Route path='/link/create' element={<LinkCreatePage />} />
            <Route path='/link/create/plaid' element={<PlaidSetup alreadyOnboarded />} />
            <Route path='/link/create/manual' element={<CreateManualLinkPage />} />
            <Route path='/logout' element={<LogoutPage />} />
            <Route path='/plaid/oauth-return' element={<OauthReturn />} />
            <Route path='/subscription' element={<SubscriptionPage />} />
            <Route path='/account/subscribe' element={<Navigate replace to='/' />} />
            <Route path='/account/subscribe/after' element={<AfterCheckoutPage />} />
            <Route path='/setup' element={<Navigate replace to='/' />} />
            <Route path='/password/reset' element={<Navigate replace to='/' />} />
            <Route path='/register' element={<Navigate replace to='/' />} />
            <Route path='/login' element={<Navigate replace to='/' />} />
            <Route path='/login/multifactor' element={<Navigate replace to='/' />} />
            <Route index path='/' element={<RedirectToBank />} />
          </RoutesImpl>
        </div>
      </div>
    </MobileSidebarContextProvider>
  );
}

function BudgetingLayout(): JSX.Element {
  return (
    <Fragment>
      <BudgetingSidebar className='hidden lg:flex' />
      <div className='min-w-0 flex flex-col grow'>
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
