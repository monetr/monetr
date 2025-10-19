import type { Meta, StoryObj } from '@storybook/react';

import MonetrWrapper, { BankView } from '@monetr/interface/pages/app';

import TransactionDetails from './details';

// import GetAPIFixtures from 'stories/apiFixtures';

const meta: Meta<typeof TransactionDetails> = {
  title: 'New UI/Transaction',
  component: TransactionDetails,
  parameters: {
    // msw: {
    //   handlers: [
    //     ...GetAPIFixtures(),
    //   ],
    // },
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
