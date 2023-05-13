import { StoryObj, Meta } from '@storybook/react';
import LoginPage from './login-new';

const meta: Meta<typeof LoginPage> = {
  title: 'Login',
  component: LoginPage,
}

export default meta;

export const Default: StoryObj<typeof LoginPage> = {
  name: 'Login View',
  args: {
    requests: [
      {
        method: 'GET',
        path: '/api/config',
        status: 200,
        response: {
          allowForgotPassword: true,
        }
      }
    ]
  }
}
