import { Meta, StoryObj } from '@storybook/react';

import RegisterPage from './register-new';

import { rest } from 'msw';

const meta: Meta<typeof RegisterPage> = {
  title: 'Pages/Authentication/Register',
  component: RegisterPage,
};

export default meta;

export const Default: StoryObj<typeof RegisterPage> = {
  name: 'Default',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: true,
            allowSignUp: true,
            requireBetaCode: false,
          }));
        }),
      ],
    },
  },
};

export const WithReCAPTCHA: StoryObj<typeof RegisterPage> = {
  name: 'With ReCAPTCHA',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: true,
            allowSignUp: true,
            ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
            verifyRegister: true,
          }));
        }),
      ],
    },
  },
};

export const WithBetaCode: StoryObj<typeof RegisterPage> = {
  name: 'Require Beta Code',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: true,
            allowSignUp: true,
            requireBetaCode: true,
          }));
        }),
      ],
    },
  },
};
