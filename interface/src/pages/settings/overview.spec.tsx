import { waitFor } from '@testing-library/react';

import SettingsOverview from '@monetr/interface/pages/settings/overview';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('overview settings page', () => {
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

  it('will show a verified email', async () => {
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: null,
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      mfaPending: false,
      trialingUntil: null,
      defaultCurrency: 'USD',
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

    const world = testRenderer(<SettingsOverview />, { initialRoute: '/settings/overview' });

    await waitFor(() => expect(world.getByTestId('settings-overview-email')).toHaveTextContent('example@example.com'));
    await waitFor(() => expect(world.getByTestId('settings-overview-email-verified')).toBeVisible());
  });

  it('will not show verified without a verified at date', async () => {
    // The server only considers the email verified if both of these are set, so the badge should follow that too
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: null,
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      mfaPending: false,
      trialingUntil: null,
      defaultCurrency: 'USD',
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
          emailVerifiedAt: null,
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

    const world = testRenderer(<SettingsOverview />, { initialRoute: '/settings/overview' });

    await waitFor(() => expect(world.getByTestId('settings-overview-email')).toHaveTextContent('example@example.com'));
    await waitFor(() => expect(world.queryByTestId('settings-overview-email-verified')).not.toBeInTheDocument());
  });
});
