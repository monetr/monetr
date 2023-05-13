import { StoryObj, Meta } from '@storybook/react';
import LoginPage from './login-new';

const meta: Meta<typeof LoginPage> = {
  title: 'Login',
  component: LoginPage,
}

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
        }
      }
    ]
  }
}

export const NoSignup: StoryObj<typeof LoginPage> = {
  name: 'No Signup',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: true,
          allowSignUp: false,
        }
      }
    ]
  }
}

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
        }
      }
    ]
  }
}
