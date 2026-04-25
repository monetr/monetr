import { act } from 'react';
import { rs } from '@rstest/core';
import { endOfMonth, endOfToday, startOfToday } from 'date-fns';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import SettingsBilling from '@monetr/interface/pages/settings/billing';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

const locationAssignMock = rs.fn();

// jsdom 26 makes window.location non-configurable. To mock location.assign,
// we spy on jsdom's internal implementation via its symbol property.
const implSymbol = Reflect.ownKeys(window.location).find(i => typeof i === 'symbol');
if (!implSymbol) {
  throw new Error('jsdom implementation symbol not found on window.location');
}

describe('billing settings page', () => {
  let mockFetch: FetchMock;
  let assignSpy: ReturnType<typeof rs.spyOn>;

  beforeAll(() => {
    assignSpy = rs.spyOn((window.location as any)[implSymbol], 'assign').mockImplementation(locationAssignMock);
  });

  beforeEach(() => {
    mockFetch = new FetchMock();
  });

  afterEach(() => {
    mockFetch.reset();
  });

  afterAll(() => {
    mockFetch.restore();
    assignSpy.mockRestore();
  });

  it('will show a trial subscription', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      requireLegalName: false,
      requirePhoneNumber: false,
      verifyLogin: false,
      verifyRegister: false,
      verifyEmailAddress: true,
      verifyForgotPassword: false,
      allowSignUp: true,
      allowForgotPassword: true,
      longPollPlaidSetup: true,
      requireBetaCode: false,
      initialPlan: {
        price: 499,
      },
      billingEnabled: true,
      iconsEnabled: true,
      plaidEnabled: true,
      manualEnabled: true,
      uploadsEnabled: true,
      release: '',
      revision: '',
      buildType: 'development',
      buildTime: '2025-01-07T19:17:19Z',
    });
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: null,
      hasSubscription: false,
      isActive: true,
      isSetup: true,
      isTrialing: true,
      mfaPending: false,
      trialingUntil: endOfToday().toISOString(),
      user: {
        userId: 'user_01jh111mq7ev2wvnnxxn5etgn5',
        loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
        login: {
          loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          passwordResetAt: null,
          isEmailVerified: true,
          emailVerifiedAt: '2025-01-07T18:39:50.227236Z',
          totpEnabledAt: null,
        },
        accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
        account: {
          accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
          timezone: 'America/Chicago',
          locale: 'en_US',
          trialEndsAt: '2025-01-08T18:39:46.406975Z',
          createdAt: '2025-01-07T18:39:46.40702Z',
        },
        role: 'owner',
      },
    });
    mockFetch.onPost('/api/billing/create_checkout').reply(200, {
      url: 'http://localhost/bogus/portal',
    });

    const world = testRenderer(<SettingsBilling />, { initialRoute: '/settings/billing' });

    await waitFor(() => expect(world.getByTestId('billing-subscription-trialing')).toBeVisible());
    await waitFor(() => expect(world.getByText('Subscribe Early')).toBeVisible());

    await act(() => userEvent.click(world.getByTestId('billing-subscribe')));

    await waitFor(() => expect(locationAssignMock).toHaveBeenCalledWith('http://localhost/bogus/portal'));
  });

  it('will show an active subscription', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      requireLegalName: false,
      requirePhoneNumber: false,
      verifyLogin: false,
      verifyRegister: false,
      verifyEmailAddress: true,
      verifyForgotPassword: false,
      allowSignUp: true,
      allowForgotPassword: true,
      longPollPlaidSetup: true,
      requireBetaCode: false,
      initialPlan: {
        price: 499,
      },
      billingEnabled: true,
      iconsEnabled: true,
      plaidEnabled: true,
      manualEnabled: true,
      uploadsEnabled: true,
      release: '',
      revision: '',
      buildType: 'development',
      buildTime: '2025-01-07T19:17:19Z',
    });
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: endOfMonth(endOfToday()).toISOString(),
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      mfaPending: false,
      trialingUntil: startOfToday().toISOString(),
      user: {
        userId: 'user_01jh111mq7ev2wvnnxxn5etgn5',
        loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
        login: {
          loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          passwordResetAt: null,
          isEmailVerified: true,
          emailVerifiedAt: '2025-01-07T18:39:50.227236Z',
          totpEnabledAt: null,
        },
        accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
        account: {
          accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
          timezone: 'America/Chicago',
          locale: 'en_US',
          subscriptionActiveUntil: endOfMonth(endOfToday()).toISOString(),
          subscriptionStatus: 'active',
          trialEndsAt: startOfToday().toISOString(),
          createdAt: '2025-01-07T18:39:46.40702Z',
        },
        role: 'owner',
      },
    });
    mockFetch.onGet('/api/billing/portal').reply(200, {
      url: 'http://localhost/bogus/portal',
    });

    const world = testRenderer(<SettingsBilling />, { initialRoute: '/settings/billing' });

    await waitFor(() => expect(world.getByTestId('billing-subscription-active')).toBeVisible());
    await waitFor(() => expect(world.getByText('Manage Your Subscription')).toBeVisible());

    await act(() => userEvent.click(world.getByTestId('billing-subscribe')));

    await waitFor(() => expect(locationAssignMock).toHaveBeenCalledWith('http://localhost/bogus/portal'));
  });

  it('will show an expired subscription', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      requireLegalName: false,
      requirePhoneNumber: false,
      verifyLogin: false,
      verifyRegister: false,
      verifyEmailAddress: true,
      verifyForgotPassword: false,
      allowSignUp: true,
      allowForgotPassword: true,
      longPollPlaidSetup: true,
      requireBetaCode: false,
      initialPlan: {
        price: 499,
      },
      billingEnabled: true,
      iconsEnabled: true,
      plaidEnabled: true,
      manualEnabled: true,
      uploadsEnabled: true,
      release: '',
      revision: '',
      buildType: 'development',
      buildTime: '2025-01-07T19:17:19Z',
    });
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: startOfToday().toISOString(),
      hasSubscription: true,
      isActive: false,
      isSetup: true,
      isTrialing: false,
      mfaPending: false,
      trialingUntil: startOfToday().toISOString(),
      user: {
        userId: 'user_01jh111mq7ev2wvnnxxn5etgn5',
        loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
        login: {
          loginId: 'lgn_01jh111mq6hfhm750wsy3p897k',
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          passwordResetAt: null,
          isEmailVerified: true,
          emailVerifiedAt: '2025-01-07T18:39:50.227236Z',
          totpEnabledAt: null,
        },
        accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
        account: {
          accountId: 'acct_01jh111mq7ev2wvnnxxjex24x3',
          timezone: 'America/Chicago',
          locale: 'en_US',
          subscriptionActiveUntil: startOfToday().toISOString(),
          subscriptionStatus: 'active',
          trialEndsAt: startOfToday().toISOString(),
          createdAt: '2025-01-07T18:39:46.40702Z',
        },
        role: 'owner',
      },
    });

    const world = testRenderer(<SettingsBilling />, { initialRoute: '/settings/billing' });

    await waitFor(() => expect(world.getByTestId('billing-subscription-expired')).toBeVisible());
    await waitFor(() => expect(world.getByText('Manage Your Subscription')).toBeVisible());
  });
});
