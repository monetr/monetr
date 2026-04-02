import { act } from 'react';

import FetchMock from '@monetr/interface/testutils/fetchMock';
import useLogout from '@monetr/interface/hooks/useLogout';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('logout', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
  });
  afterEach(() => {
    mockFetch.reset();
  });
  afterAll(() => { mockFetch.restore(); });

  it('will logout successfully', async () => {
    mockFetch.onGet('/api/authentication/logout').reply(200);

    const {
      result: { current: logout },
    } = testRenderHook(useLogout, { initialRoute: '/' });

    expect(mockFetch.history.get).toHaveLength(0);

    await act(() => {
      return logout();
    });

    // Make sure that we did make the API call.
    expect(mockFetch.history.get).toHaveLength(1);
    expect(mockFetch.history.get[0]).toMatchObject({ url: '/api/authentication/logout' });
  });
});
