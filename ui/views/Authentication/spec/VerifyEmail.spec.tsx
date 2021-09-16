import { render } from '@testing-library/react';
import axios, { AxiosResponse } from 'axios';
import mockAxios from 'jest-mock-axios';
import { VerifyEmail } from 'views/Authentication/VerifyEmail';
import { createMemoryHistory, MemoryHistory, createLocation } from 'history';
import { match } from 'react-router';
import React from 'react';
import HttpStatus from 'http-status-codes';

const path = `/verify/email`;

describe('VerifyEmail', () => {
  let history: MemoryHistory;

  beforeEach(() => {
    history = createMemoryHistory();
    window.alert = jest.fn();
    history.push = jest.fn();
  });

  it('will render without token', () => {
    mockAxios.post.mockResolvedValue(null);
    const newMatch: match<object> = {
      isExact: false,
      path,
      url: path,
      params: {},
    };
    const location = createLocation(newMatch.url);

    const verifyEmail = render(<VerifyEmail
      history={ history }
      location={ location }
      match={ newMatch }
    />);

    // Make sure that if a token is not specified in the URL; that we do not try to make a request to verify that token.
    expect(mockAxios.post).not.toHaveBeenCalled();
    // The user should be presented with an alert telling them that the link is not valid.
    expect(window.alert).toHaveBeenCalledWith('Email verification link is not valid.');
    // After that, the user should be redirected to the login page.
    expect(history.push).toHaveBeenCalledWith('/login');
    // Always make sure that all of our assertions have run properly.
    expect.assertions(3);
  });

  it('will render with an invalid token', async () => {
    const response: AxiosResponse = {
      config: undefined,
      headers: undefined,
      data: {
        error: 'Invalid token provided.'
      },
      status: HttpStatus.BAD_REQUEST,
      statusText: HttpStatus.BAD_REQUEST.toString(),
    }
    mockAxios.post.mockRejectedValueOnce({
      response,
    });

    const token = 'iAmAnInvalidToken'
    // This time we can include an arbitrary string as the token. We don't do any token parsing on the client side, we
    // assume that if the token is there it is good enough to send to the server.
    const pathWithToken = `${ path }?token=${ token }`;

    const newMatch: match<object> = {
      isExact: false,
      path: pathWithToken,
      url: pathWithToken,
      params: {},
    };
    const location = createLocation(newMatch.url);

    const verifyEmail = await render(<VerifyEmail
      history={ history }
      location={ location }
      match={ newMatch }
    />);

    // Make sure that if a token is provided in the URL, that we send that token off to the server.
    await expect(mockAxios.post).toHaveBeenCalledWith('/authentication/verify', {
      token,
    })
    // If the server responds with an error, make sure we display that error.
    expect(window.alert).toHaveBeenCalledWith(response.data.error);
    // We should pretty much always redirect the user back to the login URL by default. Same goes for this scenario.
    expect(history.push).toHaveBeenCalledWith('/login');
    // Always make sure that all of our assertions have run properly.
    expect.assertions(3);
  });

  it('will render with an invalid token and a custom nextUrl', async () => {
    const response: AxiosResponse = {
      config: undefined,
      headers: undefined,
      data: {
        error: 'Invalid token provided.',
        nextUrl: '/login?message=Bad%20link'
      },
      status: HttpStatus.BAD_REQUEST,
      statusText: HttpStatus.BAD_REQUEST.toString(),
    }
    mockAxios.post.mockRejectedValueOnce({
      response,
    });

    const token = 'iAmAnInvalidToken'
    // This time we can include an arbitrary string as the token. We don't do any token parsing on the client side, we
    // assume that if the token is there it is good enough to send to the server.
    const pathWithToken = `${ path }?token=${ token }`;

    const newMatch: match<object> = {
      isExact: false,
      path: pathWithToken,
      url: pathWithToken,
      params: {},
    };
    const location = createLocation(newMatch.url);

    const verifyEmail = await render(<VerifyEmail
      history={ history }
      location={ location }
      match={ newMatch }
    />);

    // Make sure that if a token is provided in the URL, that we send that token off to the server.
    await expect(mockAxios.post).toHaveBeenCalledWith('/authentication/verify', {
      token,
    })
    // If the server responds with an error, make sure we display that error.
    expect(window.alert).toHaveBeenCalledWith(response.data.error);
    // Make sure that if the response includes a nextUrl, that we properly redirect to that.
    expect(history.push).toHaveBeenCalledWith(response.data.nextUrl);
    // Always make sure that all of our assertions have run properly.
    expect.assertions(3);
  });

  it('will render with a valid token', async () => {
    const response: AxiosResponse = {
      config: undefined,
      headers: undefined,
      data: {
        message: 'Your email address has been verified, please login.',
      },
      status: HttpStatus.OK,
      statusText: HttpStatus.OK.toString(),
    }
    mockAxios.post.mockResolvedValue(response);

    const token = 'iAmAValidToken'
    const pathWithToken = `${ path }?token=${ token }`;

    const newMatch: match<object> = {
      isExact: false,
      path: pathWithToken,
      url: pathWithToken,
      params: {},
    };
    const location = createLocation(newMatch.url);

    const verifyEmail = await render(<VerifyEmail
      history={ history }
      location={ location }
      match={ newMatch }
    />);

    // Make sure that if a token is provided in the URL, that we send that token off to the server.
    await expect(mockAxios.post).toHaveBeenCalledWith('/authentication/verify', {
      token,
    })
    // If the server responds with an error, make sure we display that error.
    expect(window.alert).toHaveBeenCalledWith(response.data.message);
    // We should pretty much always redirect the user back to the login URL by default. Same goes for this scenario.
    expect(history.push).toHaveBeenCalledWith('/login');
    // Always make sure that all of our assertions have run properly.
    expect.assertions(3);
  });

  it('will render with a valid token and ignore other params', async () => {
    const response: AxiosResponse = {
      config: undefined,
      headers: undefined,
      data: {
        message: 'Your email address has been verified, please login.',
      },
      status: HttpStatus.OK,
      statusText: HttpStatus.OK.toString(),
    }
    mockAxios.post.mockResolvedValue(response);

    const token = 'iAmAValidToken'
    const pathWithToken = `${ path }?token=${ token }&not_token=abc1234`;

    const newMatch: match<object> = {
      isExact: false,
      path: pathWithToken,
      url: pathWithToken,
      params: {},
    };
    const location = createLocation(newMatch.url);

    const verifyEmail = await render(<VerifyEmail
      history={ history }
      location={ location }
      match={ newMatch }
    />);

    // Make sure that if a token is provided in the URL, that we send that token off to the server.
    await expect(mockAxios.post).toHaveBeenCalledWith('/authentication/verify', {
      token,
    })
    // If the server responds with an error, make sure we display that error.
    expect(window.alert).toHaveBeenCalledWith(response.data.message);
    // We should pretty much always redirect the user back to the login URL by default. Same goes for this scenario.
    expect(history.push).toHaveBeenCalledWith('/login');
    // Always make sure that all of our assertions have run properly.
    expect.assertions(3);
  });

  it('will render with a valid token and nextUrl', async () => {
    const response: AxiosResponse = {
      config: undefined,
      headers: undefined,
      data: {
        message: 'Your email address has been verified, please login.',
        nextUrl: '/login?email=test@test.com'
      },
      status: HttpStatus.OK,
      statusText: HttpStatus.OK.toString(),
    }
    mockAxios.post.mockResolvedValue(response);

    const token = 'iAmAValidToken'
    const pathWithToken = `${ path }?token=${ token }`;

    const newMatch: match<object> = {
      isExact: false,
      path: pathWithToken,
      url: pathWithToken,
      params: {},
    };
    const location = createLocation(newMatch.url);

    const verifyEmail = await render(<VerifyEmail
      history={ history }
      location={ location }
      match={ newMatch }
    />);

    // Make sure that if a token is provided in the URL, that we send that token off to the server.
    await expect(mockAxios.post).toHaveBeenCalledWith('/authentication/verify', {
      token,
    })
    // If the server responds with an error, make sure we display that error.
    expect(window.alert).toHaveBeenCalledWith(response.data.message);
    // Make sure that when the nextUrl is specified, that it does respect that.
    expect(history.push).toHaveBeenCalledWith(response.data.nextUrl);
    // Always make sure that all of our assertions have run properly.
    expect.assertions(3);
  });
});