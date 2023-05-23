import { Meta, StoryObj } from '@storybook/react';

import LoginPage from './login-new';

const meta: Meta<typeof LoginPage> = {
  title: 'Pages/Authentication/Login',
  component: LoginPage,
};

export default meta;

export const Default: StoryObj<typeof LoginPage> = {
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
        },
      },
    ],
  },
};

export const WithReCAPTCHA: StoryObj<typeof LoginPage> = {
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
          verifyLogin: true,
        },
      },
    ],
  },
};

export const NoSignup: StoryObj<typeof LoginPage> = {
  name: 'No Sign Up',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: true,
          allowSignUp: false,
        },
      },
    ],
  },
};

export const NoForgotPassword: StoryObj<typeof LoginPage> = {
  name: 'No Forgot Password',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: false,
          allowSignUp: true,
        },
      },
    ],
  },
};
