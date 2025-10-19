import { waitFor } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import testRenderer from '@monetr/interface/testutils/renderer';

import FundingDetails from './details';

describe('funding schedule details view', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will render with adjusted weekend', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
      activeUntil: '2024-09-26T00:31:38Z',
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      trialingUntil: null,
      defaultCurrency: 'USD',
      user: {
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        loginId: 'lgn_01hym36d96ze86vz5g7883vcwg',
        login: {
          loginId: 'lgn_01hym36d96ze86vz5g7883vcwg',
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          passwordResetAt: null,
          isEmailVerified: true,
          emailVerifiedAt: '2022-09-25T00:24:25.976514Z',
          totpEnabledAt: null,
        },
        accountId: 'acct_01hk84dchvxvjgp7cgap818c82',
        account: {
          accountId: 'acct_01hk84dchvxvjgp7cgap818c82',
          timezone: 'America/Chicago',
          locale: 'en_US',
          subscriptionActiveUntil: '2024-09-26T00:31:38Z',
          subscriptionStatus: 'active',
          trialEndsAt: null,
          createdAt: '2024-01-03T17:02:23.290914Z',
        },
      },
    });
    mockAxios.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
      availableBalance: 48635,
      currentBalance: 48635,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      officialName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      currency: 'USD',
      status: 'active',
      lastUpdated: '2023-07-02T04:22:52.48118Z',
    });

    mockAxios
      .onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
        dateStarted: '2023-02-28T06:00:00Z',
        description: '15th and last day of every month',
        estimatedDeposit: null,
        excludeWeekends: true,
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx', // 3,
        lastRecurrence: '2023-09-29T05:00:00Z',
        name: "Elliot's Contribution",
        nextRecurrence: '2023-10-13T05:00:00Z',
        nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        waitForDeposit: false,
      });

    const world = testRenderer(<FundingDetails />, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding/fund_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('funding-details-date-picker')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('funding-schedule-weekend-notice')).toBeVisible());
  });

  it('will render without adjusted weekend', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
      activeUntil: '2024-09-26T00:31:38Z',
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      trialingUntil: null,
      defaultCurrency: 'USD',
      user: {
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        loginId: 'lgn_01hym36d96ze86vz5g7883vcwg',
        login: {
          loginId: 'lgn_01hym36d96ze86vz5g7883vcwg',
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          passwordResetAt: null,
          isEmailVerified: true,
          emailVerifiedAt: '2022-09-25T00:24:25.976514Z',
          totpEnabledAt: null,
        },
        accountId: 'acct_01hk84dchvxvjgp7cgap818c82',
        account: {
          accountId: 'acct_01hk84dchvxvjgp7cgap818c82',
          timezone: 'America/Chicago',
          locale: 'en_US',
          subscriptionActiveUntil: '2024-09-26T00:31:38Z',
          subscriptionStatus: 'active',
          trialEndsAt: null,
          createdAt: '2024-01-03T17:02:23.290914Z',
        },
      },
    });
    mockAxios.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
      availableBalance: 48635,
      currentBalance: 48635,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      officialName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      currency: 'USD',
      status: 'active',
      lastUpdated: '2023-07-02T04:22:52.48118Z',
    });

    mockAxios
      .onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
        dateStarted: '2023-02-28T06:00:00Z',
        description: '15th and last day of every month',
        estimatedDeposit: null,
        excludeWeekends: false,
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx', // 3,
        lastRecurrence: '2023-09-30T05:00:00Z',
        name: "Elliot's Contribution",
        nextRecurrence: '2023-10-15T05:00:00Z',
        nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        waitForDeposit: false,
      });

    const world = testRenderer(<FundingDetails />, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding/fund_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('funding-details-date-picker')).toBeVisible());
    await waitFor(() => expect(world.queryByTestId('funding-schedule-weekend-notice')).not.toBeInTheDocument());
  });
});
