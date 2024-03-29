import React from 'react';
import * as reactRouter from 'react-router-dom';
import { act, fireEvent, waitFor } from '@testing-library/react';
import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';

import Login from '@monetr/interface/pages/login';
import testRenderer from '@monetr/interface/testutils/renderer';

import { afterAll, afterEach, beforeEach, describe, expect, it, jest, mock, test } from 'bun:test';

// const mockUseNavigate = mock((_url: string) => { });
// // jest.mock('react-router-dom', () => ({
// //   __esModule: true,
// //   ...jest.requireActual('react-router-dom'),
// //   useNavigate: () => mockUseNavigate,
// // }));
//
// mock.module('react-router-dom', () => {
//   return {
//     // __esModule: true,
//     // ...require('react-router-dom'),
//     useNavigate: () => mockUseNavigate,
//   };
// });

const mockUseNavigate = jest.fn((_url: string) => { });
mock.module(reactRouter, () => ({
  // __esModule: true,
  // ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockUseNavigate,
}));


describe('login page', () => {
  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(axios);
  });
  afterEach(() => {
    mock.reset();
  });
  afterAll(() => mock.restore());

  it('will render with default options', async () => {
    mock.onGet('/api/config').reply(200, {
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
    mock.onGet('/api/config').reply(200, {
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
    mock.onGet('/api/config').reply(200, {
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
    mock.onGet('/api/config').reply(200, {
      allowForgotPassword: false,
      allowSignUp: false,
      verifyLogin: false,
    });

    mock.onPost('/api/authentication/login').reply(200, {
      isActive: true,
    });

    const world = testRenderer(<Login />, { initialRoute: '/login' });

    await waitFor(() => expect(world.getByTestId('login-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('login-submit')).toBeVisible());

    await waitFor(() => expect(world.queryByTestId('login-forgot')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('login-signup')).not.toBeInTheDocument());

    // act(() => {
    fireEvent.change(world.getByTestId('login-email'), { target: { value: 'test@test.com' } });
    fireEvent.change(world.getByTestId('login-password'), { target: { value: 'password' } });
    fireEvent.click(world.getByTestId('login-submit'));
    // world.getByTestId('login-submit').click();
    // });

    console.log('CALLS', mockUseNavigate.mock.calls);
    // When we login we should be redirected to this route.
    // await waitFor(() => expect(mockUseNavigate).toBeCalledWith('/'));
  });

  test('will submit login and require subscription', async () => {
    mock.onGet('/api/config').reply(200, {
      allowForgotPassword: false,
      allowSignUp: false,
      verifyLogin: false,
    });

    mock.onPost('/api/authentication/login').reply(200, {
      isActive: false,
      nextUrl: '/account/subscribe',
    });

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
    // await waitFor(() => expect(mockUseNavigate).toBeCalledWith('/account/subscribe'));
  });
});
