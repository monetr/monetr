import type { Meta, StoryObj } from '@storybook/react';

import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import LoginPage from '@monetr/interface/pages/login';

const meta: Meta<typeof LoginPage> = {
  title: 'Pages/Authentication/Login',
  component: LoginPage,
};

export default meta;

export const Default: StoryObj<typeof LoginPage> = {
  name: 'Default',
  decorators: [
    (Story, _) => {
      const mockAxios = new MockAdapter(monetrClient);
      mockAxios.onGet('/api/config').reply(200, {
        allowForgotPassword: true,
        allowSignUp: true,
        verifyLogin: false,
      });
      mockAxios.onPost('/api/authentication/login').reply(403, {
        error: 'Invalid credentials provided!',
      });

      return <Story />;
    },
  ],
};

export const WithReCAPTCHA: StoryObj<typeof LoginPage> = {
  name: 'With ReCAPTCHA',
  decorators: [
    (Story, _) => {
      const mockAxios = new MockAdapter(monetrClient);
      mockAxios.onGet('/api/config').reply(200, {
        allowForgotPassword: true,
        allowSignUp: true,
        ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
        verifyLogin: true,
      });
      mockAxios.onPost('/api/authentication/login').reply(403, {
        error: 'Invalid credentials provided!',
      });

      return <Story />;
    },
  ],
};

export const NoSignup: StoryObj<typeof LoginPage> = {
  name: 'No Sign Up',
  decorators: [
    (Story, _) => {
      const mockAxios = new MockAdapter(monetrClient);
      mockAxios.onGet('/api/config').reply(200, {
        allowForgotPassword: true,
        allowSignUp: false,
      });
      mockAxios.onPost('/api/authentication/login').reply(403, {
        error: 'Invalid credentials provided!',
      });

      return <Story />;
    },
  ],
};

export const NoForgotPassword: StoryObj<typeof LoginPage> = {
  name: 'No Forgot Password',
  decorators: [
    (Story, _) => {
      const mockAxios = new MockAdapter(monetrClient);
      mockAxios.onGet('/api/config').reply(200, {
        allowForgotPassword: false,
        allowSignUp: true,
      });
      mockAxios.onPost('/api/authentication/login').reply(403, {
        error: 'Invalid credentials provided!',
      });

      return <Story />;
    },
  ],
};
