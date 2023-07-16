/* eslint-disable max-len */
import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import MonetrWrapper, { BankView, ExpensesView, TransactionsView } from './new';


import GetAPIFixtures from 'stories/apiFixtures';
import Monetr from 'monetr';

const meta: Meta<typeof MonetrWrapper> = {
  title: 'New UI',
  component: MonetrWrapper,
  parameters: {
    msw: {
      handlers: [
        ...GetAPIFixtures(),
      ],
    },
  },
};

export default meta;

export const Transactions: StoryObj<typeof MonetrWrapper> = {
  name: 'Transactions',
  render: () => (
    <Monetr />
  ),
  parameters: {
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/transactions',
    },
  },
};

export const Expenses: StoryObj<typeof MonetrWrapper> = {
  name: 'Expenses',
  render: () => (
    <Monetr />
  ),
  parameters: {
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/expenses',
      routeParams: {
        bankId: '12',
      },
    },
  },
};
