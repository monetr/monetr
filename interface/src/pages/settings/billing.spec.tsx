import { act } from 'react';
import { endOfMonth, endOfToday, startOfToday } from 'date-fns';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import SettingsBilling from '@monetr/interface/pages/settings/billing';
import testRenderer from '@monetr/interface/testutils/renderer';

const oldWindowLocation = window.location;
const locationAssignMock = jest.fn();

describe('billing settings page', () => {
  let mockAxios: MockAdapter;

  beforeAll(() => {
    delete window.location;
    // @ts-expect-error
    window.location = Object.defineProperties(
      {},
      {
        ...Object.getOwnPropertyDescriptors(oldWindowLocation),
        assign: {
          configurable: true,
          value: locationAssignMock,
        },
      },
    );
  });

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });

  afterEach(() => {
    mockAxios.reset();
  });

  afterAll(() => {
    mockAxios.restore();
    window.location = oldWindowLocation;
  });

  it('will show a trial subscription', async () => {
    mockAxios.onGet('/api/config').reply(200, {
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
    mockAxios.onGet('/api/users/me').reply(200, {
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
    mockAxios.onPost('/api/billing/create_checkout').reply(200, {
      url: 'http://localhost/bogus/portal',
    });

    const world = testRenderer(<SettingsBilling />, { initialRoute: '/settings/billing' });

    await waitFor(() => expect(world.getByTestId('billing-subscription-trialing')).toBeVisible());
    await waitFor(() => expect(world.getByText('Subscribe Early')).toBeVisible());

    await act(() => userEvent.click(world.getByTestId('billing-subscribe')));

    await waitFor(() => expect(locationAssignMock).toHaveBeenCalledWith('http://localhost/bogus/portal'));
  });

  it('will show an active subscription', async () => {
    mockAxios.onGet('/api/config').reply(200, {
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
    mockAxios.onGet('/api/users/me').reply(200, {
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
    mockAxios.onGet('/api/billing/portal').reply(200, {
      url: 'http://localhost/bogus/portal',
    });

    const world = testRenderer(<SettingsBilling />, { initialRoute: '/settings/billing' });

    await waitFor(() => expect(world.getByTestId('billing-subscription-active')).toBeVisible());
    await waitFor(() => expect(world.getByText('Manage Your Subscription')).toBeVisible());

    await act(() => userEvent.click(world.getByTestId('billing-subscribe')));

    await waitFor(() => expect(locationAssignMock).toHaveBeenCalledWith('http://localhost/bogus/portal'));
  });

  it('will show an expired subscription', async () => {
    mockAxios.onGet('/api/config').reply(200, {
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
    mockAxios.onGet('/api/users/me').reply(200, {
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
