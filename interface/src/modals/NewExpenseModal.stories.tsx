import { Meta, StoryObj } from '@storybook/react';
import { rest } from 'msw';

import NewExpenseModal, { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';

import GetAPIFixtures from 'stories/apiFixtures';

const meta: Meta<typeof NewExpenseModal> = {
  title: 'New UI/Modals/Expenses/Create',
  component: NewExpenseModal,
};

export default meta;

export const Default: StoryObj<typeof NewExpenseModal> = {
  render: () => {
    showNewExpenseModal();
    return null;
  },
  parameters: {
    msw: {
      handlers: [
        ...GetAPIFixtures(),
        rest.post('/api/bank_accounts/12/spending', (_req, res, ctx) => {
          return res(ctx.json({
            'bankAccountId': 12,
            'currentAmount': 0,
            'dateCreated': '2023-05-14T20:01:47.09268Z',
            'dateStarted': '2023-06-10T05:00:00Z',
            'description': 'Every month on the 10th',
            'fundingScheduleId': 3,
            'isBehind': false,
            'isPaused': false,
            'lastRecurrence': '2023-07-01T05:00:00Z',
            'name': 'Dummy Expense Created',
            'nextContributionAmount': 120,
            'nextRecurrence': '2023-08-01T05:00:00Z',
            'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
            'spendingId': 100,
            'spendingType': 0,
            'targetAmount': 500,
            'usedAmount': 0,
          }));
        }),
      ],
    },
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/expenses',
      routeParams: {
        bankAccountId: 12,
      },
    },
  },
};

export const NoFundingSchedule: StoryObj<typeof NewExpenseModal> = {
  render: () => {
    showNewExpenseModal();
    return null;
  },
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/bank_accounts/12', (_req, res, ctx) => {
          return res(ctx.delay(200), ctx.json({
            'bankAccountId': 12,
            'linkId': 4,
            'availableBalance': 48635,
            'currentBalance': 48635,
            'mask': '2982',
            'name': 'Mercury Checking',
            'originalName': 'Mercury Checking',
            'officialName': 'Mercury Checking',
            'accountType': 'depository',
            'accountSubType': 'checking',
            'status': 'active',
            'lastUpdated': '2023-07-02T04:22:52.48118Z',
          }));
        }),
        rest.get('/api/bank_accounts/12/funding_schedules', (_req, res, ctx) => {
          return res(ctx.delay(200), ctx.json([]));
        }),
      ],
    },
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/expenses',
      routeParams: {
        bankAccountId: 12,
      },
    },
  },
};
