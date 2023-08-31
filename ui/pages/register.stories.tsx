import { Meta, StoryObj } from '@storybook/react';

import RegisterPage, { RegisterSuccessful } from 'pages/register';

import { rest } from 'msw';
import React from 'react';

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
        rest.post('/api/authentication/register', async (req, res, ctx) => {
          const body = await req.json();
          switch (body['lastName'].toString().toLower()) {
            case 'already':
              return res(ctx.status(400), ctx.json({
                error: 'email already in use',
                code: 'EMAIL_IN_USE',
              }));
            case 'verify':
              return res(ctx.json({
                message: 'A verification email has been sent to your email address, please verify your email.',
                requireVerification: true,
              }));
            case 'servererror':
              return res(ctx.status(500), ctx.json({
                error: 'An internal error occurred.',
              }));
            case 'bill':
              return res(ctx.json({
                nextUrl: '/account/subscribe',
                isActive: false,
                requireVerification: false,
              }));
          }

          return res(ctx.json({
            isActive: true,
            nextUrl: '/setup',
            requireVerification: false,
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

export const Successful: StoryObj<typeof RegisterPage> = {
  name: 'Successful',
  render: () => <RegisterSuccessful />,
};
