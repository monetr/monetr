import { act } from '@testing-library/react-hooks';
import { rest } from 'msw';

import useLogin from '@monetr/interface/hooks/useLogin';
import testRenderHook from '@monetr/interface/testutils/hooks';
import { server } from '@monetr/interface/testutils/server';

const mockUseNavigate = jest.fn((_url: string) => {});
jest.mock('react-router-dom', () => ({
  __esModule: true,
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockUseNavigate,
}));

describe('login', () => {
  it('will authenticate successfully', async () => {
    server.use(
      rest.post('/api/authentication/login', (_req, res, ctx) => {
        return res(ctx.json({
          isActive: false,
          nextUrl: '/account/subscribe',
        }));
      }),
    );

    const { result: { current: login } } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // Make sure we end up navigating to the url returned by the login endpoint.
    expect(mockUseNavigate).toBeCalledWith('/account/subscribe');
  });

  it('will navigate without a next url', async () => {
    server.use(
      rest.post('/api/authentication/login', (_req, res, ctx) => {
        return res(ctx.json({
          isActive: true,
        }));
      }),
    );

    const { result: { current: login } } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When the login endpoint does not return a next url, navigate to an index route.
    expect(mockUseNavigate).toBeCalledWith('/');
  });

  it('will require a password reset', async () => {
    server.use(
      rest.post('/api/authentication/login', (_req, res, ctx) => {
        return res(ctx.status(428), ctx.json({
          code: 'PASSWORD_CHANGE_REQUIRED',
          resetToken: 'abc123',
        }));
      }),
    );

    const { result: { current: login } } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When the login endpoint returns a password change required error; then make sure we navigate to the password
    // reset page.
    expect(mockUseNavigate).toBeCalledWith('/password/reset', {
      state: {
        'message': 'You are required to change your password before authenticating.',
        'token': 'abc123',
      },
    });
  });

  it('email has not been verified', async () => {
    server.use(
      rest.post('/api/authentication/login', (_req, res, ctx) => {
        return res(ctx.status(428), ctx.json({
          code: 'EMAIL_NOT_VERIFIED',
        }));
      }),
    );

    const { result: { current: login } } = testRenderHook(useLogin, { initialRoute: '/login' });

    await act(() => {
      return login({
        email: 'test@test.com',
        password: 'password',
      });
    });

    // When our email is not verified, make sure we navigate to the resend page.
    expect(mockUseNavigate).toBeCalledWith('/verify/email/resend', {
      state: {
        'emailAddress': 'test@test.com',
      },
    });
  });
});
