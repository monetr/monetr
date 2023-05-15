import { Meta, StoryObj } from '@storybook/react';

import MSidebar from '.';

const meta: Meta<typeof MSidebar> = {
  title: 'Pages/Templates/Sidebar',
  component: MSidebar,
};

export default meta;

export const Default: StoryObj<typeof MSidebar> = {
  name: 'Default',
  args: {
    open: true,
    // @ts-ignore
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          billingEnabled: true,
        },
      },
    ],
  },
};

export const BillingDisabled: StoryObj<typeof MSidebar> = {
  name: 'Billing Disabled',
  args: {
    open: true,
    // @ts-ignore
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          billingEnabled: false,
        },
      },
    ],
  },
};
