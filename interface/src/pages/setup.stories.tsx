import { Meta, StoryObj } from '@storybook/react';
import { rest } from 'msw';

import SetupPage from './setup';

const meta: Meta<typeof SetupPage> = {
  title: 'Pages/Setup',
  component: SetupPage,
};

export default meta;

export const Default: StoryObj<typeof SetupPage> = {
  name: 'Default',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/plaid/link/token/new', (_req, res, ctx) => {
          return res(ctx.json({
            linkToken: 'link-sandbox-a21289fc-1363-4632-b877-33bacc6dc069',
          }));
        }),
      ],
    },
  },
};

export const FailBeforePlaid: StoryObj<typeof SetupPage> = {
  name: 'Fail Before Plaid',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/plaid/link/token/new', (_req, res, ctx) => {
          return res(ctx.status(500), ctx.json({
            error: 'Unable to create a plaid link token',
          }));
        }),
      ],
    },
  },
};

export const ManualEnabled: StoryObj<typeof SetupPage> = {
  name: 'Manual Enabled',
  args: {
    manualEnabled: true,
  },
};
