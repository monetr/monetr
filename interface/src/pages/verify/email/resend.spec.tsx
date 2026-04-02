import { waitFor } from '@testing-library/react';

import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

import ResendVerificationPage from './resend';

describe('resend verification email', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
  });
  afterEach(() => {
    mockFetch.reset();
  });

  afterAll(() => {
    mockFetch.restore();
  });

  it('will render without ReCAPTCHA', () => {
    mockFetch.onGet('/api/config').reply(200, {
      ReCAPTCHAKey: null,
    });

    const world = testRenderer(<ResendVerificationPage />, { initialRoute: '/verify/email/resend' });

    expect(world.queryByTestId('resend-email')).toBeVisible();
    expect(world.queryByTestId('resend-captcha')).not.toBeInTheDocument();
    expect(world.queryByTestId('resend-email-excluded')).toBeVisible();
    expect(world.queryByTestId('resend-email-included')).not.toBeInTheDocument();
  });

  it('will render with ReCAPTCHA', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      ReCAPTCHAKey: '6LfL3vcgAAAAALlJNxvUPdgrbzH_ca94YTCqso6L',
    });

    const world = testRenderer(<ResendVerificationPage />, { initialRoute: '/verify/email/resend' });

    expect(world.queryByTestId('resend-email')).toBeVisible();
    expect(world.queryByTestId('resend-email-excluded')).toBeVisible();
    expect(world.queryByTestId('resend-email-included')).not.toBeInTheDocument();
    await waitFor(() => expect(world.queryByTestId('resend-captcha')).toBeVisible());
  });

  it('will render with provided email', async () => {
    mockFetch.onGet('/api/config').reply(200, {
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
