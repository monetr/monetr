import { rs } from '@rstest/core';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import Register from '@monetr/interface/pages/register';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('register page', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
    mockEnqueueSnackbar.mockReset();
  });
  afterEach(() => {
    mockFetch.reset();
  });

  afterAll(() => {
    mockFetch.restore();
  });

  it('will render with default options', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowSignUp: true,
    });

    const world = testRenderer(<Register />, { initialRoute: '/register' });

    await waitFor(() => expect(world.getByTestId('register-first-name')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-last-name')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-confirm-password')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('register-submit')).toBeVisible());
  });

  it('will toast when there are too many requests', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowSignUp: true,
      verifyRegister: false,
    });

    mockFetch.onPost('/api/authentication/register').reply(429);

    const world = testRenderer(<Register />, { initialRoute: '/register' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('register-submit')).toBeVisible());

    await user.type(world.getByTestId('register-first-name'), 'Test');
    await user.type(world.getByTestId('register-last-name'), 'User');
    await user.type(world.getByTestId('register-email'), 'test@test.com');
    await user.type(world.getByTestId('register-password'), 'password');
    await user.type(world.getByTestId('register-confirm-password'), 'password');
    await user.click(world.getByTestId('register-submit'));

    // We should toast a friendly message telling the user they have been rate limited.
    await waitFor(() =>
      expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Too many requests, please try again in a few minutes', {
        variant: 'error',
        disableWindowBlurListener: true,
      }),
    );
  });

  it('will toast the message from the api for other errors', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowSignUp: true,
      verifyRegister: false,
    });

    // This one is a normal error coming from monetr itself so it does have a JSON body with a message. Make sure we did
    // not break that path when we added the rate limit handling above, we should still surface the api's message.
    mockFetch.onPost('/api/authentication/register').reply(400, {
      error: 'That email address is already in use.',
    });

    const world = testRenderer(<Register />, { initialRoute: '/register' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('register-submit')).toBeVisible());

    await user.type(world.getByTestId('register-first-name'), 'Test');
    await user.type(world.getByTestId('register-last-name'), 'User');
    await user.type(world.getByTestId('register-email'), 'test@test.com');
    await user.type(world.getByTestId('register-password'), 'password');
    await user.type(world.getByTestId('register-confirm-password'), 'password');
    await user.click(world.getByTestId('register-submit'));

    await waitFor(() =>
      expect(mockEnqueueSnackbar).toHaveBeenCalledWith('That email address is already in use.', {
        variant: 'error',
        disableWindowBlurListener: true,
      }),
    );
  });
});
