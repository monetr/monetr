import { Meta, StoryObj } from '@storybook/react';

import LoginPage from './login-new';

import { rest } from 'msw';

const meta: Meta<typeof LoginPage> = {
  title: 'Pages/Authentication/Login',
  component: LoginPage,
};

export default meta;

export const Default: StoryObj<typeof LoginPage> = {
  name: 'Default',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: true,
            allowSignUp: true,
          }));
        }),
        rest.post('/api/authentication/login', (_req, res, ctx) => {
          return res(
            ctx.delay(500),
            ctx.status(403),
            ctx.json({
              error: 'Invalid credentials provided!',
            }),
          );
        }),
      ],
    },
  },
};

export const WithReCAPTCHA: StoryObj<typeof LoginPage> = {
  name: 'With ReCAPTCHA',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: true,
            allowSignUp: true,
            ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
            verifyLogin: true,
          }));
        }),
      ],
    },
  },
};

export const NoSignup: StoryObj<typeof LoginPage> = {
  name: 'No Sign Up',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: true,
            allowSignUp: false,
          }));
        }),
      ],
    },
  },
};

export const NoForgotPassword: StoryObj<typeof LoginPage> = {
  name: 'No Forgot Password',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: false,
            allowSignUp: true,
          }));
        }),
      ],
    },
  },
};
