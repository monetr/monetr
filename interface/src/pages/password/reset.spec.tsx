import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import PasswordResetNew from '@monetr/interface/pages/password/reset';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('reset password page', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
  });
  afterEach(() => {
    mockFetch.reset();
    window.history.replaceState({}, '', '/');
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will keep the token after cleaning it out of the url', async () => {
    // The page reads the token out of the query string and then immediately strips it back out of the URL so that it is
    // not left sitting in the address bar. Wouter patches `history.replaceState` in order to make `useSearch` reactive,
    // so that cleanup used to wipe the token out from under the page before it could ever be used.
    window.history.replaceState({}, '', '/password/reset?token=abc123');

    mockFetch.onPost('/api/authentication/reset').reply(200, {});

    const world = testRenderer(<PasswordResetNew />, { browserLocation: true, initialRoute: '/password/reset' });
    const user = userEvent.setup();

    await waitFor(() => expect(world.getByTestId('reset-password')).toBeVisible());

    // The token should no longer be visible in the URL, but the page must still be holding onto it.
    expect(window.location.search).toBe('');

    await user.type(world.getByTestId('reset-password'), 'notATerriblePassword');
    await user.type(world.getByTestId('reset-verify-password'), 'notATerriblePassword');
    await user.click(world.getByTestId('reset-submit'));

    await waitFor(() => {
      const resetPost = mockFetch.history.post?.find(entry => entry.url === '/api/authentication/reset');
      expect(resetPost?.data).toMatchObject({ token: 'abc123', password: 'notATerriblePassword' });
    });
  });

  it('will show a different message when a password change is required', async () => {
    window.history.replaceState({}, '', '/password/reset?token=abc123&reason=password_change_required');

    const world = testRenderer(<PasswordResetNew />, { browserLocation: true, initialRoute: '/password/reset' });

    await waitFor(() =>
      expect(world.getByText('You are required to change your password before authenticating.')).toBeVisible(),
    );
  });

  it('will send the user back to login without a token', async () => {
    window.history.replaceState({}, '', '/password/reset');

    testRenderer(<PasswordResetNew />, { browserLocation: true, initialRoute: '/password/reset' });

    await waitFor(() => expect(window.location.pathname).toBe('/login'));
    expect(mockFetch.history.post ?? []).toHaveLength(0);
  });
});
