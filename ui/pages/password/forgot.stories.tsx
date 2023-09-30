import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import ForgotPasswordPage, { ForgotPasswordComplete } from 'pages/password/forgot';

const meta: Meta<typeof ForgotPasswordPage> = {
  title: 'Pages/Authentication/Forgot Password',
  component: ForgotPasswordPage,
};

export default meta;

export const Default: StoryObj<typeof ForgotPasswordPage> = {
  name: 'Default',
};

export const Complete: StoryObj<typeof ForgotPasswordComplete> = {
  name: 'Complete',
  render: () => <ForgotPasswordComplete />,
};
