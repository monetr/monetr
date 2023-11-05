import React from 'react';
import { act, fireEvent, waitFor } from '@testing-library/react';
import { rest } from 'msw';

import Login from '@monetr/interface/pages/login';
import testRenderer from '@monetr/interface/testutils/renderer';
import { server } from '@monetr/interface/testutils/server';

const mockUseNavigate = jest.fn((_url: string) => { });
jest.mock('react-router-dom', () => ({
  __esModule: true,
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockUseNavigate,
}));

describe('login page', () => {
  it('will render with default options', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: true,
          allowSignUp: true,
          verifyLogin: false,
        }));
      }),
    );

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-signup')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-forgot')).toBeVisible());
  });

  it('without signup', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: true,
          allowSignUp: false,
          verifyLogin: false,
        }));
      }),
    );

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-forgot')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());
  });

  it('without forgot password', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: false,
          allowSignUp: false,
          verifyLogin: false,
        }));
      }),
    );

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-forgot')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());
  });

  it('will submit login', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: false,
          allowSignUp: false,
          verifyLogin: false,
        }));
      }),
      rest.post('/api/authentication/login', (_req, res, ctx) => {
        return res(ctx.json({
          isActive: true,
        }));
      }),
    );

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-forgot')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());

    act(() => {
      fireEvent.change(world.getByTestId('login-email'), { target: { value: 'test@test.com' } });
      fireEvent.change(world.getByTestId('login-password'), { target: { value: 'password' } });
      world.getByTestId('login-submit').click();
    });

    // When we login we should be redirected to this route.
    await waitFor(() => expect(mockUseNavigate).toBeCalledWith('/'));
  });

  it('will submit login and require subscription', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: false,
          allowSignUp: false,
          verifyLogin: false,
        }));
      }),
      rest.post('/api/authentication/login', (_req, res, ctx) => {
        return res(ctx.json({
          isActive: false,
          nextUrl: '/account/subscribe',
        }));
      }),
    );

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-forgot')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());

    act(() => {
      fireEvent.change(world.getByTestId('login-email'), { target: { value: 'test@test.com' } });
      fireEvent.change(world.getByTestId('login-password'), { target: { value: 'password' } });
      world.getByTestId('login-submit').click();
    });

    // When we login we should be redirected to the subscribe page
    await waitFor(() => expect(mockUseNavigate).toBeCalledWith('/account/subscribe'));
  });
});
