import type { Meta, StoryObj } from '@storybook/react';

import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';

const meta: Meta<typeof BudgetingSidebar> = {
  title: 'Layout/Budgeting Sidebar',
  component: BudgetingSidebar,
};

export default meta;

export const Default: StoryObj<typeof BudgetingSidebar> = {
  name: 'Default',
  decorators: [
    (Story, _) => {
      const mockAxios = new MockAdapter(monetrClient);
      mockAxios.onGet('/api/config').reply(200, {
        requireLegalName: true,
        requirePhoneNumber: true,
        verifyLogin: false,
        verifyRegister: false,
        verifyEmailAddress: true,
        verifyForgotPassword: false,
        allowSignUp: true,
        allowForgotPassword: true,
        longPollPlaidSetup: true,
        requireBetaCode: true,
        initialPlan: {
          price: 199,
        },
        billingEnabled: true,
        iconsEnabled: true,
        plaidEnabled: true,
        manualEnabled: false,
        release: '0.17.16',
        revision: '8df5505b7e5273f061d90ddf19e4c1cfca2b4f4f',
        buildType: 'binary',
        buildTime: '2024-08-28T02:42:56Z',
      });
      mockAxios.onGet('/api/users/me').reply(200, {
        activeUntil: '2024-09-26T00:31:38Z',
        hasSubscription: true,
        isActive: true,
        isSetup: true,
        isTrialing: false,
        trialingUntil: null,
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
      mockAxios.onGet('/api/links').reply(200, [
        {
          linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
          linkType: 1,
          plaidLink: {
            products: ['transactions'],
            status: 2,
            expirationDate: null,
            newAccountsAvailable: false,
            institutionId: 'ins_116794',
            institutionName: 'Mercury',
            lastManualSync: '2024-07-06T12:59:09.51222Z',
            lastSuccessfulUpdate: '2024-08-29T12:00:01.176597Z',
            lastAttemptedUpdate: '2024-08-29T12:00:01.17665Z',
            updatedAt: '2024-03-19T06:17:32.335106Z',
            createdAt: '2022-09-25T02:08:40.758642Z',
            createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
          },
          institutionName: 'Mercury',
          description: null,
          createdAt: '2022-09-25T02:08:40.758642Z',
          createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
          updatedAt: '2024-03-19T06:17:32.335106Z',
          deletedAt: null,
        },
      ]);
      mockAxios.onGet('/api/bank_accounts').reply(200, [
        {
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
          availableBalance: 47986,
          currentBalance: 47986,
          mask: '2982',
          name: 'Mercury Checking',
          originalName: 'Mercury Checking',
          accountType: 'depository',
          accountSubType: 'checking',
          status: 'active',
          lastUpdated: '2024-08-27T08:53:48.555368Z',
          createdAt: '2022-09-25T02:08:40.758642Z',
          updatedAt: '2024-03-19T06:17:32.335106Z',
        },
      ]);
      mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
        plaidBankAccount: {
          name: 'Mercury Checking',
          officialName: 'Mercury Checking',
          mask: '2982',
          availableBalance: 47986,
          currentBalance: 47986,
          limitBalance: null,
          createdAt: '2024-03-19T15:15:10.31132Z',
          createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        },
        availableBalance: 47986,
        currentBalance: 47986,
        mask: '2982',
        name: 'Mercury Checking',
        originalName: 'Mercury Checking',
        accountType: 'depository',
        accountSubType: 'checking',
        status: 'active',
        lastUpdated: '2024-08-27T08:53:48.555368Z',
        createdAt: '2022-09-25T02:08:40.758642Z',
        updatedAt: '2024-03-19T06:17:32.335106Z',
      });
      mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules').reply(200, [
        {
          fundingScheduleId: 'fund_01hym37k3kj4ghv67nfx7vkvr0',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          name: "Elliot's Contribution",
          description: '15th and last day of every month',
          ruleset: 'DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          excludeWeekends: true,
          waitForDeposit: false,
          estimatedDeposit: 22000,
          lastRecurrence: '2024-12-13T06:00:00Z',
          nextRecurrence: '2024-12-31T06:00:00Z',
          nextRecurrenceOriginal: '2024-12-31T06:00:00Z',
        },
      ]);
      mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/balances').reply(200, {
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        current: 70879,
        available: 70879,
        free: 10000,
        expenses: 18630,
        goals: 42249,
      });
      return <Story />;
    },
  ],
  parameters: {
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions',
      routeParams: {
        bankAccountId: 12,
        spendingId: 9999,
      },
    },
  },
  render: () => <BudgetingSidebar className='w-full border-none' />,
};
