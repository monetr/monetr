import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import ExpenseDetails from './details';

import Monetr from 'monetr';
import GetAPIFixtures from 'stories/apiFixtures';

const meta: Meta<typeof ExpenseDetails> = {
  title: 'New UI/Expense',
  component: ExpenseDetails,
  parameters: {
    msw: {
      handlers: [
        ...GetAPIFixtures(),
      ],
    },
  },
};

export default meta;

export const NotFound: StoryObj<typeof ExpenseDetails> = {
  name: 'Not Found',
  render: () => (
    <Monetr />
  ),
  parameters: {
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/expenses/9999/details',
      routeParams: {
        bankAccountId: 12,
        spendingId: 9999,
      },
    },
  },
};

export const ExpenseDetailCloudProduction: StoryObj<typeof ExpenseDetails> = {
  name: 'Expense Detail (Cloud Production)',
  render: () => (
    <Monetr />
  ),
  parameters: {
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/expenses/191/details',
      routeParams: {
        bankAccountId: 12,
        spendingId: 191,
      },
    },
  },
};

export const ExpenseDetailGitLab: StoryObj<typeof ExpenseDetails> = {
  name: 'Expense Detail (GitLab)',
  render: () => (
    <Monetr />
  ),
  parameters: {
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/expenses/63/details',
      routeParams: {
        bankAccountId: 12,
        spendingId: 63,
      },
    },
  },
};
