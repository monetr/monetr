import { Meta, StoryObj } from '@storybook/react';
import { rest } from 'msw';

import NewFundingModal, { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';

const meta: Meta<typeof NewFundingModal> = {
  title: 'New UI/Modals/Funding/Create',
  component: NewFundingModal,
};

export default meta;

export const Default: StoryObj<typeof NewFundingModal> = {
  render: () => {
    showNewFundingModal();
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
      ],
    },
    reactRouter: {
      routePath: '/*',
      browserPath: '/bank/12/funding',
      routeParams: {
        bankAccountId: 12,
      },
    },
  },
};
