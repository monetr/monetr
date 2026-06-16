import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import ForgotPasswordNew from '@monetr/interface/pages/password/forgot';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('forgot password page', () => {
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

  it('will render with default options', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: true,
      verifyForgotPassword: false,
      proofOfWorkEnabled: false,
    });

    const world = testRenderer(<ForgotPasswordNew />, { initialRoute: '/password/forgot' });

    await waitFor(() => expect(world.getByTestId('forgot-email')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('forgot-submit')).toBeVisible());
  });

  it('will send a forgot password request', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: true,
      verifyForgotPassword: false,
      proofOfWorkEnabled: false,
    });
    mockFetch.onPost('/api/authentication/forgot').reply(200, {});

    const world = testRenderer(<ForgotPasswordNew />, { initialRoute: '/password/forgot' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('forgot-email')).toBeVisible());
    await user.type(world.getByTestId('forgot-email'), 'test@test.com');
    await user.click(world.getByTestId('forgot-submit'));

    // Once the request goes through we show the "check your email" screen.
    await waitFor(() => expect(world.getByText('Check your email')).toBeVisible());

    const forgotPost = mockFetch.history.post?.find(entry => entry.url === '/api/authentication/forgot');
    expect(forgotPost?.data).toMatchObject({ email: 'test@test.com' });
  });

  it('will include the proof of work solution when enabled', async () => {
    mockFetch.onGet('/api/config').reply(200, {
      allowForgotPassword: true,
      verifyForgotPassword: false,
      proofOfWorkEnabled: true,
    });

    // A difficulty of 0 means the solver returns a nonce of 0 immediately.
    mockFetch.onPost('/api/authentication/challenge').reply(200, {
      challenge: 'x',
      difficulty: 0,
    });
    mockFetch.onPost('/api/authentication/forgot').reply(200, {});

    const world = testRenderer(<ForgotPasswordNew />, { initialRoute: '/password/forgot' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('forgot-email')).toBeVisible());
    await user.type(world.getByTestId('forgot-email'), 'test@test.com');
    await user.click(world.getByTestId('forgot-submit'));

    await waitFor(() => expect(world.getByText('Check your email')).toBeVisible());

    // The forgot request should carry the challenge and the nonce we solved.
    const forgotPost = mockFetch.history.post?.find(entry => entry.url === '/api/authentication/forgot');
    expect(forgotPost?.data).toMatchObject({ email: 'test@test.com', challenge: 'x', nonce: 0 });
  });
});
