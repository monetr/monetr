import { waitFor } from '@testing-library/react';

import Register from '@monetr/interface/pages/register';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('register page', () => {
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
});
