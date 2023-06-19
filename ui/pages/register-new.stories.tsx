import { Meta, StoryObj } from '@storybook/react';

import RegisterPage from './register-new';

const meta: Meta<typeof RegisterPage> = {
  title: 'Pages/Authentication/Register',
  component: RegisterPage,
};

export default meta;

export const Default: StoryObj<typeof RegisterPage> = {
  name: 'Default',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: true,
          allowSignUp: true,
          requireBetaCode: false,
        },
      },
    ],
  },
};

export const WithReCAPTCHA: StoryObj<typeof RegisterPage> = {
  name: 'With ReCAPTCHA',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: true,
          allowSignUp: true,
          ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
          verifyRegister: true,
        },
      },
    ],
  },
};

export const WithBetaCode: StoryObj<typeof RegisterPage> = {
  name: 'Require Beta Code',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: true,
          allowSignUp: true,
          requireBetaCode: true,
        },
      },
    ],
  },
};
