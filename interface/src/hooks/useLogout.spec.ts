import { act } from 'react';

import useLogout from '@monetr/interface/hooks/useLogout';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('logout', () => {
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

  it('will logout successfully', async () => {
    mockFetch.onGet('/api/authentication/logout').reply(200);

    const {
      result: { current: logout },
    } = testRenderHook(useLogout, { initialRoute: '/' });

    expect(mockFetch.history.get).toHaveLength(0);

    await act(() => {
      return logout();
    });

    // Make sure that we did make the API call. The history is keyed by method so typescript thinks the get bucket might
    // be undefined, pull it into a local and assert its there before we poke at the first entry.
    const getHistory = mockFetch.history.get;
    expect(getHistory).toBeDefined();
    expect(getHistory).toHaveLength(1);
    expect(getHistory?.[0]).toMatchObject({ url: '/api/authentication/logout' });
  });
});
