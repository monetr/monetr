import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import ExpensesPage from './expenses-new';

import MLayout from 'components/MLayout';
import { rest } from 'msw';

const meta: Meta<typeof ExpensesPage> = {
  title: 'Pages/Expenses',
  component: ExpensesPage,
};

export default meta;

export const Default: StoryObj<typeof ExpensesPage> = {
  name: 'Default',
  render: () => (
    <MLayout>
      <ExpensesPage />
    </MLayout>
  ),
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({}));
        }),
        rest.get('/api/links', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'linkId': 4,
              'linkType': 1,
              'plaidInstitutionId': 'ins_116794',
              'plaidNewAccountsAvailable': false,
              'linkStatus': 2,
              'expirationDate': null,
              'institutionName': 'Mercury',
              'description': null,
              'createdAt': '2022-09-25T02:08:40.758642Z',
              'createdByUserId': 1,
              'updatedAt': '2023-05-13T09:01:24.952957Z',
              'lastManualSync': '2023-05-02T19:56:34.953077Z',
              'lastSuccessfulUpdate': '2023-05-13T09:01:24.952559Z',
            },
          ]));
        }),
        rest.get('/api/bank_accounts', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'bankAccountId': 1,
              'linkId': 4,
              'availableBalance': 70479,
              'currentBalance': 70479,
              'mask': '1234',
              'name': 'Mercury Checking',
              'originalName': 'Mercury Checking',
              'officialName': 'Mercury Checking',
              'accountType': 'depository',
              'accountSubType': 'checking',
              'status': 'active',
              'lastUpdated': '2023-05-13T09:01:24.147095Z',
            },
          ]));
        }),
        rest.get('/api/bank_accounts/1/funding_schedules', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'fundingScheduleId': 3,
              'bankAccountId': 12,
              'name': 'Elliot\'s Contribution',
              'description': '15th and last day of every month',
              'rule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
              'excludeWeekends': true,
              'waitForDeposit': false,
              'estimatedDeposit': null,
              'lastOccurrence': '2023-05-15T05:00:00Z',
              'nextOccurrence': '2023-05-31T05:00:00Z',
              'dateStarted': '2023-02-28T06:00:00Z',
            },
          ]));
        }),
        rest.get('/api/bank_accounts/1/spending', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'spendingId': 136,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'BuildKite',
              'description': 'Every month on the 27th',
              'targetAmount': 1500,
              'currentAmount': 0,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=27',
              'lastRecurrence': '2023-05-27T05:00:00Z',
              'nextRecurrence': '2023-05-27T05:00:00Z',
              'nextContributionAmount': 1500,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:42:11Z',
              'dateStarted': '2023-02-28T06:00:00Z',
            }, {
              'spendingId': 137,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Sentry',
              'description': 'Every month on the 25th',
              'targetAmount': 2900,
              'currentAmount': 0,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=25',
              'lastRecurrence': '2023-05-25T05:00:00Z',
              'nextRecurrence': '2023-05-25T05:00:00Z',
              'nextContributionAmount': 2900,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:42:41Z',
              'dateStarted': '2023-03-25T05:00:00Z',
            }, {
              'spendingId': 201,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 1,
              'name': 'Rainy Day',
              'targetAmount': 10000,
              'currentAmount': 31600,
              'usedAmount': 0,
              'recurrenceRule': null,
              'lastRecurrence': '2022-12-31T06:00:00Z',
              'nextRecurrence': '2022-12-31T06:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-29T16:32:58Z',
              'dateStarted': '2022-12-31T06:00:00Z',
            }, {
              'spendingId': 192,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Cloud Staging',
              'description': 'Every month on the 1st',
              'targetAmount': 10000,
              'currentAmount': 1874,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-06-01T05:00:00Z',
              'nextRecurrence': '2023-06-01T05:00:00Z',
              'nextContributionAmount': 4063,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-07T15:09:32Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            }, {
              'spendingId': 189,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Google Voice',
              'description': 'Every month on the 1st',
              'targetAmount': 1366,
              'currentAmount': 21,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-06-01T05:00:00Z',
              'nextRecurrence': '2023-06-01T05:00:00Z',
              'nextContributionAmount': 672,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-02T14:11:24Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            }, {
              'spendingId': 208,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Google Domains ($12)',
              'description': 'Every year on the 29th of January',
              'targetAmount': 1200,
              'currentAmount': 347,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=YEARLY;INTERVAL=1;BYMONTH=1;BYMONTHDAY=29',
              'lastRecurrence': null,
              'nextRecurrence': '2024-01-29T06:00:00Z',
              'nextContributionAmount': 50,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2023-01-30T22:16:20Z',
              'dateStarted': '2024-01-29T06:00:00Z',
            }, {
              'spendingId': 171,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'ngrok',
              'description': 'Every year on the 26th of June',
              'targetAmount': 6000,
              'currentAmount': 5184,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=YEARLY;INTERVAL=1;BYMONTH=6;BYMONTHDAY=26',
              'lastRecurrence': null,
              'nextRecurrence': '2023-06-25T05:00:00Z',
              'nextContributionAmount': 204,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-06-28T15:59:10Z',
              'dateStarted': '2023-06-25T05:00:00Z',
            }, {
              'spendingId': 134,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Freshbooks',
              'description': 'Every month on the 10th',
              'targetAmount': 1700,
              'currentAmount': 0,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10',
              'lastRecurrence': '2023-06-10T05:00:00Z',
              'nextRecurrence': '2023-06-10T05:00:00Z',
              'nextContributionAmount': 850,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:40:46Z',
              'dateStarted': '2023-03-10T06:00:00Z',
            }, {
              'spendingId': 138,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'GitHub',
              'description': 'Every month on the 19th',
              'targetAmount': 2600,
              'currentAmount': 2600,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=19',
              'lastRecurrence': '2023-05-19T05:00:00Z',
              'nextRecurrence': '2023-05-19T05:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:43:04Z',
              'dateStarted': '2023-03-19T05:00:00Z',
            }, {
              'spendingId': 191,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Cloud Production',
              'description': 'Every month on the 1st',
              'targetAmount': 28000,
              'currentAmount': 3841,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-06-01T05:00:00Z',
              'nextRecurrence': '2023-06-01T05:00:00Z',
              'nextContributionAmount': 12079,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-07T15:09:16Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            }, {
              'spendingId': 135,
              'bankAccountId': 1,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'G-Suite ($12)',
              'description': 'Every month on the 1st',
              'targetAmount': 1200,
              'currentAmount': 0,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-06-01T05:00:00Z',
              'nextRecurrence': '2023-06-01T05:00:00Z',
              'nextContributionAmount': 600,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:41:18Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            },
          ]));
        }),
      ],
    },
  },
};

export const EmptyState: StoryObj<typeof ExpensesPage> = {
  name: 'Empty State',
  render: () => (
    <MLayout>
      <ExpensesPage />
    </MLayout>
  ),
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({}));
        }),
        rest.get('/api/links', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'linkId': 4,
              'linkType': 1,
              'plaidInstitutionId': 'ins_116794',
              'plaidNewAccountsAvailable': false,
              'linkStatus': 2,
              'expirationDate': null,
              'institutionName': 'Mercury',
              'description': null,
              'createdAt': '2022-09-25T02:08:40.758642Z',
              'createdByUserId': 1,
              'updatedAt': '2023-05-13T09:01:24.952957Z',
              'lastManualSync': '2023-05-02T19:56:34.953077Z',
              'lastSuccessfulUpdate': '2023-05-13T09:01:24.952559Z',
            },
          ]));
        }),
        rest.get('/api/bank_accounts', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'bankAccountId': 1,
              'linkId': 4,
              'availableBalance': 70479,
              'currentBalance': 70479,
              'mask': '1234',
              'name': 'Mercury Checking',
              'originalName': 'Mercury Checking',
              'officialName': 'Mercury Checking',
              'accountType': 'depository',
              'accountSubType': 'checking',
              'status': 'active',
              'lastUpdated': '2023-05-13T09:01:24.147095Z',
            },
          ]));
        }),
        rest.get('/api/bank_accounts/1/funding_schedules', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'fundingScheduleId': 3,
              'bankAccountId': 12,
              'name': 'Elliot\'s Contribution',
              'description': '15th and last day of every month',
              'rule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
              'excludeWeekends': true,
              'waitForDeposit': false,
              'estimatedDeposit': null,
              'lastOccurrence': '2023-05-15T05:00:00Z',
              'nextOccurrence': '2023-05-31T05:00:00Z',
              'dateStarted': '2023-02-28T06:00:00Z',
            },
          ]));
        }),
        rest.get('/api/bank_accounts/1/spending', (_req, res, ctx) => {
          return res(ctx.json([]));
        }),
      ],
    },
  },
};
