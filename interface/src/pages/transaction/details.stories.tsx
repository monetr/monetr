/* eslint-disable max-len */
import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import TransactionDetails from './details';

import MonetrWrapper, { BankView } from 'pages/new';
import GetAPIFixtures from 'stories/apiFixtures';

const meta: Meta<typeof TransactionDetails> = {
  title: 'New UI/Transaction',
  component: TransactionDetails,
  parameters: {
    msw: {
      handlers: [
        ...GetAPIFixtures(),
      ],
    },
  },
};

export default meta;


export const NoTransaction: StoryObj<typeof TransactionDetails> = {
  name: 'No Transaction',
  render: () => (
    <MonetrWrapper>
      <BankView>
        <TransactionDetails />
      </BankView>
    </MonetrWrapper>
  ),
};
