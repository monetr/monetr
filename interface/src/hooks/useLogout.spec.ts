import { act } from '@testing-library/react-hooks';
import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';

import useLogout from '@monetr/interface/hooks/useLogout';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('logout', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(axios);
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will logout successfully', async () => {
    mockAxios.onGet('/api/authentication/logout').reply(200);

    const { result: { current: logout } } = testRenderHook(useLogout, { initialRoute: '/' });

    expect(mockAxios.history['get']).toHaveLength(0);

    await act(() => {
      return logout();
    });

    // Make sure that we did make the API call.
    expect(mockAxios.history['get']).toHaveLength(1);
    expect(mockAxios.history['get'][0]).toMatchObject({ url: '/api/authentication/logout' });
  });
});
