import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import { VerifyEmailView } from './email';

const meta: Meta<typeof VerifyEmailView> = {
  title: 'Verify Email',
};

export default meta;

export const Default: StoryObj<typeof VerifyEmailView> = {
  name: 'Default',
  render: () => <VerifyEmailView />,
};
