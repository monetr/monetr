import { rs } from '@rstest/core';
import * as reactRouterDomActual from 'react-router-dom' with { rstest: 'importActual' };

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import Login from '@monetr/interface/pages/login';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

const mockUseNavigate = rs.fn((_url: string) => {});
rs.mock('react-router-dom', () => ({
  ...reactRouterDomActual,
  useNavigate: () => mockUseNavigate,
}));

describe('login page', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
    mockUseNavigate.mockReset();
  });
  afterEach(() => {
    mockFetch.reset();
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will render with default options', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: true,
      allowSignUp: true,
      verifyLogin: false,
    });

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-signup')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-forgot')).toBeVisible());
  });

  test('without signup', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: true,
      allowSignUp: false,
      verifyLogin: false,
    });

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-forgot')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());
  });

  test('without forgot password', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: false,
      allowSignUp: false,
      verifyLogin: false,
    });

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-forgot')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());
  });

  test('will submit login', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: false,
      allowSignUp: false,
      verifyLogin: false,
    });

    mockFetch.onPost('/api/authentication/login').reply(200, {
      isActive: true,
    });

    const world = testRenderer(<Login />, { initialRoute: '/login' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-forgot')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());

    await user.type(world.getByTestId('login-email'), 'test@test.com');
    await user.type(world.getByTestId('login-password'), 'password');
    await user.click(world.getByTestId('login-submit'));

    // When we login we should be redirected to this route.
    await waitFor(() => expect(mockUseNavigate).toHaveBeenCalledWith('/'));
  });

  test('will submit login and require subscription', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: false,
      allowSignUp: false,
      verifyLogin: false,
    });

    mockFetch.onPost('/api/authentication/login').reply(200, {
      isActive: false,
      nextUrl: '/account/subscribe',
    });

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-forgot')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());

    const user = userEvent.setup();
    await user.type(world.getByTestId('login-email'), 'test@test.com');
    await user.type(world.getByTestId('login-password'), 'password');
    await user.click(world.getByTestId('login-submit'));

    // When we login we should be redirected to this route.
    await waitFor(() => expect(mockUseNavigate).toHaveBeenCalledWith('/account/subscribe'));
  });
});
