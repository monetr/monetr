import { Redirect, Route, Switch } from 'wouter';

import BudgetingLayout from '@monetr/interface/components/Layout/BudgetLayout';
import MobileSidebarContextProvider from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import SettingsLayout from '@monetr/interface/components/Layout/SettingsLayout';
import Sidebar from '@monetr/interface/components/Layout/Sidebar';
import SidebarPaddingLayout from '@monetr/interface/components/Layout/SidebarPaddingLayout';
import LunchFlowSetupAccounts from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupAccounts';
import LunchFlowSetupIntro from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupIntro';
import LunchFlowSetupSync from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupSync';
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

export default function Monetr(): JSX.Element {
  const { data: config, isLoading: configIsLoading, isError: configIsError } = useAppConfiguration();
  const { isLoading: authIsLoading, data: auth, isError: isAuthError } = useAuthentication();
  const { isLoading: linksIsLoading, data: links, isError: isLinksError } = useLinks();
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
  if (configIsError || isAuthError || isLinksError) {
    return <ConfigError />;
  }

  if (!isAuthenticated) {
    return (
      <Switch>
        <Route component={Login} path='/login' />
        <Route component={LogoutPage} path='/logout' />
        {config?.allowSignUp && <Route component={Register} path='/register' />}
        {config?.allowForgotPassword && <Route component={ForgotPassword} path='/password/forgot' />}
        <Route component={PasswordReset} path='/password/reset' />
        <Route component={VerifyEmail} path='/verify/email' />
        <Route component={ResendVerificationPage} path='/verify/email/resend' />
        <Route path='/'>
          <Redirect replace to='/login' />
        </Route>
        <Route path='*'>
          <Redirect replace to='/login' />
        </Route>
      </Switch>
    );
  }

  // If the currently authenticated user requires MFA then only allow them to access the MFA pages.
  if (auth?.mfaPending) {
    return (
      <Switch>
        <Route component={MultifactorAuthenticationPage} path='/login/multifactor' />
        <Route component={LogoutPage} path='/logout' />
        <Route path='*'>
          <Redirect replace to='/login/multifactor' />
        </Route>
      </Switch>
    );
  }

  if (!auth?.isActive) {
    return (
      <Switch>
        <Route component={LogoutPage} path='/logout' />
        <Route component={SubscribePage} path='/account/subscribe' />
        <Route component={AfterCheckoutPage} path='/account/subscribe/after' />
        <Route path='*'>
          <Redirect replace to='/account/subscribe' />
        </Route>
      </Switch>
    );
  }

  const hasAnyLinks = links?.length > 0;
  if (!hasAnyLinks) {
    return (
      <Switch>
        <Route component={LogoutPage} path='/logout' />
        <Route path='/setup'>{() => <SetupPage manualEnabled={config?.manualEnabled} />}</Route>
        <Route path='/setup/plaid'>{() => <PlaidSetup alreadyOnboarded />}</Route>
        <Route component={SetupManualLinkPage} path='/setup/manual' />
        <Route component={LunchFlowSetupIntro} path='/setup/lunchflow' />
        <Route component={LunchFlowSetupAccounts} path='/setup/lunchflow/:lunchFlowLinkId' />
        <Route component={OauthReturn} path='/plaid/oauth-return' />
        <Route path='/account/subscribe/after'>
          <Redirect replace to='/setup' />
        </Route>
        <Route path='*'>
          <Redirect replace to='/setup' />
        </Route>
      </Switch>
    );
  }

  return (
    <MobileSidebarContextProvider>
      <Sidebar />
      <Switch>
        <Route path='/bank/:bankAccountId/*'>
          <BudgetingLayout>
            <Switch>
              <Route component={BankAccountSettingsPage} path='/bank/:bankAccountId/settings' />
              <Route component={Transactions} path='/bank/:bankAccountId/transactions' />
              <Route component={TransactionDetails} path='/bank/:bankAccountId/transactions/:transactionId/details' />
              <Route component={Expenses} path='/bank/:bankAccountId/expenses' />
              <Route component={ExpenseDetails} path='/bank/:bankAccountId/expenses/:spendingId/details' />
              <Route component={Goals} path='/bank/:bankAccountId/goals' />
              <Route component={GoalDetails} path='/bank/:bankAccountId/goals/:spendingId/details' />
              <Route component={Funding} path='/bank/:bankAccountId/funding' />
              <Route component={FundingDetails} path='/bank/:bankAccountId/funding/:fundingId/details' />
              <Route path='*'>
                <Redirect replace to='/' />
              </Route>
            </Switch>
          </BudgetingLayout>
        </Route>
        <Route path='/setup'>
          <Redirect replace to='/' />
        </Route>
        <Route path='/password/reset'>
          <Redirect replace to='/' />
        </Route>
        <Route path='/register'>
          <Redirect replace to='/' />
        </Route>
        <Route path='/login/multifactor'>
          <Redirect replace to='/' />
        </Route>
        <Route path='/login'>
          <Redirect replace to='/' />
        </Route>
        <Route component={RedirectToBank} path='/' />
        <Route path='*'>
          <SidebarPaddingLayout>
            <Switch>
              <Route path='/settings'>
                <Redirect replace to='/settings/overview' />
              </Route>
              <Route path='/settings/:rest*'>
                <SettingsLayout>
                  <Switch>
                    <Route component={SettingsOverview} path='/settings/overview' />
                    <Route component={SettingsSecurity} path='/settings/security' />
                    {config?.billingEnabled && <Route component={SettingsBilling} path='/settings/billing' />}
                    <Route component={SettingsAbout} path='/settings/about' />
                    <Route path='*'>
                      <Redirect replace to='/settings/overview' />
                    </Route>
                  </Switch>
                </SettingsLayout>
              </Route>
              <Route component={LinkDetails} path='/link/:linkId/details' />
              <Route component={LinkCreatePage} path='/link/create' />
              <Route path='/link/create/plaid'>{() => <PlaidSetup alreadyOnboarded />}</Route>
              <Route component={CreateManualLinkPage} path='/link/create/manual' />
              <Route component={LunchFlowSetupIntro} path='/link/create/lunchflow' />
              <Route component={LunchFlowSetupAccounts} path='/link/create/lunchflow/:lunchFlowLinkId' />
              <Route component={LunchFlowSetupSync} path='/link/create/lunchflow/:linkId/sync' />
              <Route component={LogoutPage} path='/logout' />
              <Route component={OauthReturn} path='/plaid/oauth-return' />
              <Route component={SubscriptionPage} path='/subscription' />
              <Route path='/account/subscribe'>
                <Redirect replace to='/' />
              </Route>
              <Route component={AfterCheckoutPage} path='/account/subscribe/after' />
              <Route component={LunchFlowSetupAccounts} path='/setup/lunchflow/:lunchFlowLinkId' />
              <Route path='*'>
                <Redirect replace to='/' />
              </Route>
            </Switch>
          </SidebarPaddingLayout>
        </Route>
      </Switch>
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
    return <Redirect replace to='/link/create' />;
  }

  const account = accounts[0];

  return <Redirect replace to={`/bank/${account.bankAccountId}/transactions`} />;
}
