import { waitFor } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import testRenderer from '@monetr/interface/testutils/renderer';

import ResendVerificationPage from './resend';

describe('resend verification email', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });

  afterAll(() => mockAxios.restore());

  it('will render without ReCAPTCHA', () => {
    mockAxios.onGet('/api/config').reply(200, {
      ReCAPTCHAKey: null,
    });

    const world = testRenderer(<ResendVerificationPage />, { initialRoute: '/verify/email/resend' });

    expect(world.queryByTestId('resend-email')).toBeVisible();
    expect(world.queryByTestId('resend-captcha')).not.toBeInTheDocument();
    expect(world.queryByTestId('resend-email-excluded')).toBeVisible();
    expect(world.queryByTestId('resend-email-included')).not.toBeInTheDocument();
  });

  it('will render with ReCAPTCHA', async () => {
    mockAxios.onGet('/api/config').reply(200, {
      ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
    });

    const world = testRenderer(<ResendVerificationPage />, { initialRoute: '/verify/email/resend' });

    expect(world.queryByTestId('resend-email')).toBeVisible();
    expect(world.queryByTestId('resend-email-excluded')).toBeVisible();
    expect(world.queryByTestId('resend-email-included')).not.toBeInTheDocument();
    await waitFor(() => expect(world.queryByTestId('resend-captcha')).toBeVisible());
  });

  it('will render with provided email', async () => {
    mockAxios.onGet('/api/config').reply(200, {
      ReCAPTCHAKey: null,
    });

    const world = testRenderer(<ResendVerificationPage />, {
      initialRoute: {
        pathname: '/verify/email/resend',
        state: {
          emailAddress: 'email@test.com',
        },
      },
    });

    await waitFor(() => {
      expect(world.queryByTestId('resend-email')).toBeVisible();
      expect(world.queryByTestId('resend-captcha')).not.toBeInTheDocument();
      expect(world.queryByTestId('resend-email-included')).toBeVisible();
      expect(world.queryByTestId('resend-email-excluded')).not.toBeInTheDocument();
    });
  });
});
