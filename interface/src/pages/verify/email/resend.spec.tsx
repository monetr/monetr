import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

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
      proofOfWorkEnabled: false,
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
      proofOfWorkEnabled: false,
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
      proofOfWorkEnabled: false,
    });

    const world = testRenderer(<ResendVerificationPage />, {
      initialRoute: `/verify/email/resend?email=${encodeURIComponent('email@test.com')}`,
    });

    await waitFor(() => {
      expect(world.queryByTestId('resend-email')).toBeVisible();
      expect(world.queryByTestId('resend-captcha')).not.toBeInTheDocument();
      expect(world.queryByTestId('resend-email-included')).toBeVisible();
      expect(world.queryByTestId('resend-email-excluded')).not.toBeInTheDocument();
    });
  });

  it('will include the proof of work solution when enabled', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      ReCAPTCHAKey: null,
      proofOfWorkEnabled: true,
    });

    // Difficulty 0 means the solver returns immediately.
    mockFetch.onPost('/api/authentication/challenge').reply(200, {
      challenge: 'x',
      difficulty: 0,
      ttl: 300,
    });
    mockFetch.onPost('/api/authentication/verify/resend').reply(200, {});

    const world = testRenderer(<ResendVerificationPage />, { initialRoute: '/verify/email/resend' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('resend-email')).toBeVisible());
    await user.type(world.getByTestId('resend-email'), 'test@test.com');
    await user.click(world.getByRole('button', { name: 'Resend Verification' }));

    // The resend request should carry the challenge and the nonce we solved.
    await waitFor(() => {
      const post = mockFetch.history.post?.find(entry => entry.url === '/api/authentication/verify/resend');
      expect(post?.data).toMatchObject({ challenge: 'x', nonce: 0 });
    });
  });
});
