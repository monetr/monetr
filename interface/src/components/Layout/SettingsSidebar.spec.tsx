import { endOfToday } from 'date-fns';

import { waitFor } from '@testing-library/react';

import SettingsSidebar from '@monetr/interface/components/Layout/SettingsSidebar';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('settings sidebar', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
  });
  afterEach(() => {
    mockFetch.reset();
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will show when the trial ends', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: true,
      allowSignUp: true,
      iconsEnabled: true,
      billingEnabled: true,
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
          trialEndsAt: endOfToday().toISOString(),
          createdAt: '2025-01-07T18:39:46.40702Z',
        },
        role: 'owner',
      },
    });
    mockFetch.onGet('/api/keys').reply(200, []);

    const world = testRenderer(<SettingsSidebar />, { initialRoute: '/settings/overview' });

    await waitFor(() => expect(world.getByTestId('settings-sidebar-api-keys')).toHaveTextContent('0 keys'));
    await waitFor(() => expect(world.getByTestId('settings-sidebar-billing')).toHaveTextContent('Trial ends'));
  });

  it('will hide billing when billing is disabled', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: true,
      allowSignUp: true,
      iconsEnabled: true,
      billingEnabled: false,
    });
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: null,
      hasSubscription: false,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      mfaPending: false,
      trialingUntil: null,
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
          createdAt: '2025-01-07T18:39:46.40702Z',
        },
        role: 'owner',
      },
    });
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc',
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01jh111mq7ev2wvnnxxn5etgn5',
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
    ]);

    const world = testRenderer(<SettingsSidebar />, { initialRoute: '/settings/overview' });

    // Wait for the data to load so we aren't just checking billing before the config comes back
    await waitFor(() => expect(world.getByTestId('settings-sidebar-api-keys')).toHaveTextContent('1 key'));
    await waitFor(() => expect(world.queryByTestId('settings-sidebar-billing')).not.toBeInTheDocument());
  });
});
