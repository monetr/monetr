import { act } from 'react';
import { rs } from '@rstest/core';
import * as wouterActual from 'wouter' with { rstest: 'importActual' };

import useLogin from '@monetr/interface/hooks/useLogin';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

const mockNavigate = rs.fn((_url: string) => {});
rs.mock('wouter', () => ({
  ...wouterActual,
  useLocation: () => ['/login', mockNavigate],
}));

describe('login', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
    mockNavigate.mockReset();
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
});
