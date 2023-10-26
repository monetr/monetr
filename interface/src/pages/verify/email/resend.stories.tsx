import { Meta, StoryObj } from '@storybook/react';

import ResendVerificationPage from './resend';

import { rest } from 'msw';

const meta: Meta<typeof ResendVerificationPage> = {
  title: 'Resend Verification Email',
  component: ResendVerificationPage,
};

export default meta;

export const Default: StoryObj<typeof ResendVerificationPage> = {
  name: 'Default',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            ReCAPTCHAKey: null,
          }));
        }),
      ],
    },
  },
};

export const WithReCAPTCHA: StoryObj<typeof ResendVerificationPage> = {
  name: 'With ReCAPTCHA',
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
          }));
        }),
      ],
    },
  },
};
