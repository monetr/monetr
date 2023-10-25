import React from 'react';
import { waitFor } from '@testing-library/react';

import ResendVerificationPage from './resend';

import { rest } from 'msw';
import testRenderer from 'testutils/renderer';
import { server } from 'testutils/server';

describe('resend verification email', () => {
  it('will render without ReCAPTCHA', () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          ReCAPTCHAKey: null,
        }));
      }),
    );

    const world = testRenderer(<ResendVerificationPage />, { initialRoute: '/verify/email/resend' });

    expect(world.queryByTestId('resend-email')).toBeVisible();
    expect(world.queryByTestId('resend-captcha')).not.toBeInTheDocument();
    expect(world.queryByTestId('resend-email-excluded')).toBeVisible();
    expect(world.queryByTestId('resend-email-included')).not.toBeInTheDocument();
  });

  it('will render with ReCAPTCHA', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
        }));
      }),
    );

    const world = testRenderer(<ResendVerificationPage />, { initialRoute: '/verify/email/resend' });

    expect(world.queryByTestId('resend-email')).toBeVisible();
    expect(world.queryByTestId('resend-email-excluded')).toBeVisible();
    expect(world.queryByTestId('resend-email-included')).not.toBeInTheDocument();
    await waitFor(() => expect(world.queryByTestId('resend-captcha')).toBeVisible());
  });

  it('will render with provided email', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          ReCAPTCHAKey: null,
        }));
      }),
    );

    const world = testRenderer(
      <ResendVerificationPage />,
      {
        initialRoute: {
          pathname: '/verify/email/resend',
          state: {
            emailAddress: 'email@test.com',
          },
        },
      },
    );

    expect(world.queryByTestId('resend-email')).toBeVisible();
    expect(world.queryByTestId('resend-captcha')).not.toBeInTheDocument();
    expect(world.queryByTestId('resend-email-included')).toBeVisible();
    expect(world.queryByTestId('resend-email-excluded')).not.toBeInTheDocument();
  });
});
