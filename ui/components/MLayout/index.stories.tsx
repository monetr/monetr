import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import MLayout from '.';

const meta: Meta<typeof MLayout> = {
  title: 'Pages/Templates/Layout',
  component: MLayout,
};

export default meta;

export const Default: StoryObj = {
  name: 'Default',
  render: () => (
    <MLayout>
      <div className="flex justify-center items-center h-full w-full">
        <h1>[ CONTENT ]</h1>
      </div>
    </MLayout>
  ),
  args: {
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
