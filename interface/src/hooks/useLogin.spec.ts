import { act } from 'react';
import { rs } from '@rstest/core';
import * as wouterActual from 'wouter' with { rstest: 'importActual' };

import useLogin from '@monetr/interface/hooks/useLogin';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockNavigate = rs.fn((_url: string) => {});
rs.mock('wouter', () => ({
  ...wouterActual,
  useLocation: () => ['/login', mockNavigate],
}));

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('login', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
    mockNavigate.mockReset();
    mockEnqueueSnackbar.mockReset();
  });
  afterEach(() => {
    mockFetch.reset();
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will authenticate successfully', async () => {
    mockFetch.onPost('/api/authentication/login').reply(200, {
      isActive: false,
      nextUrl: '/account/subscribe',
    });

    const {
      result: { current: login },
    } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // Make sure we end up navigating to the url returned by the login endpoint.
    expect(mockNavigate).toHaveBeenCalledWith('/account/subscribe');
  });

  it('will navigate without a next url', async () => {
    mockFetch.onPost('/api/authentication/login').reply(200, {
      isActive: true,
    });

    const {
      result: { current: login },
    } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When the login endpoint does not return a next url, navigate to an index route.
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('will require a password reset', async () => {
    mockFetch.onPost('/api/authentication/login').reply(428, {
      code: 'PASSWORD_CHANGE_REQUIRED',
      resetToken: 'abc123',
    });

    const {
      result: { current: login },
    } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When the login endpoint returns a password change required error; then make sure we navigate to the password
    // reset page.
    expect(mockNavigate).toHaveBeenCalledWith('/password/reset?token=abc123&reason=password_change_required');
  });

  it('email has not been verified', async () => {
    mockFetch.onPost('/api/authentication/login').reply(428, {
      code: 'EMAIL_NOT_VERIFIED',
    });

    const {
      result: { current: login },
    } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When our email is not verified, make sure we navigate to the resend page.
    expect(mockNavigate).toHaveBeenCalledWith('/verify/email/resend?email=test%40test.com');
  });

  it('will toast when there are too many requests', async () => {
    mockFetch.onPost('/api/authentication/login').reply(429);

    const {
      result: { current: login },
    } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // We should toast a friendly message telling the user they have been rate limited.
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Too many requests, please try again in a few minutes', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
    // And we should not navigate anywhere, the user just stays on the login page so they can try again in a bit.
    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('will not re-throw when rate limited so the login page does not double toast', async () => {
    // Same as above, a bodyless 429 from haproxy upstream.
    mockFetch.onPost('/api/authentication/login').reply(429);

    const {
      result: { current: login },
    } = testRenderHook(useLogin, { initialRoute: '/login' });

    // The hook owns this error and handles it by toasting, so the promise it returns should resolve rather than reject.
    // If it threw instead then the login page would catch it and stack its own generic "Failed to authenticate." toast
    // right on top of our rate limit one.
    let threw = false;
    await act(async () => {
      await login({
        email: 'test@test.com',
        password: 'password',
      }).catch(() => {
        threw = true;
      });
    });

    expect(threw).toBe(false);
    // Only the one rate limit toast should have been shown.
    expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1);
  });
});
