
import { Meta, StoryObj } from '@storybook/react';

import TransferModal, { showTransferModal } from './TransferModal';

import GetAPIFixtures from 'stories/apiFixtures';

const meta: Meta<typeof TransferModal> = {
  title: 'New UI/Modals/Transfer',
  component: TransferModal,
  parameters: {
    msw: {
      handlers: [
        ...GetAPIFixtures(),
      ],
    },
  },
};

export default meta;

export const NoSelection: StoryObj<typeof TransferModal> = {
  name: 'No Pre-Selection',
  render: () => {
    showTransferModal({
      initialFromSpendingId: null,
      initialToSpendingId: null,
    });
    return null;
  },
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

export const DestinationPreSelected: StoryObj<typeof TransferModal> = {
  name: 'Destination Pre-Selected',
  render: () => {
    showTransferModal({
      initialFromSpendingId: null,
      initialToSpendingId: 58,
    });
    return null;
  },
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
