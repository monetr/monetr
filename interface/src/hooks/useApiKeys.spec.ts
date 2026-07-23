import { waitFor } from '@testing-library/react';

import useApiKeys from '@monetr/interface/hooks/useApiKeys';
import ApiKey from '@monetr/interface/models/ApiKey';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('will list all api keys', () => {
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

  it('will request all api keys', async () => {
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc', // 1,
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk', // 4,
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
    ]);

    const world = testRenderHook(useApiKeys, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.isLoading).toBeTruthy());
    await waitFor(() => expect(world.result.current.isFetching).toBeTruthy());
    await waitFor(() => expect(world.result.current.data).toBeDefined());
    await waitFor(() => expect(world.result.current.data).toHaveLength(1));
  });

  it('will hydrate the api keys into ApiKey models', async () => {
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc', // 1,
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk', // 4,
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
    ]);

    const world = testRenderHook(useApiKeys, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.data).toHaveLength(1));

    // The hook should give us real ApiKey instances back, not just the raw json, otherwise the parsed dates everything
    // downstream relies on wont exist.
    const apiKey = world.result.current.data?.[0];
    expect(apiKey).toBeInstanceOf(ApiKey);
    expect(apiKey?.apiKeyId).toBe('key_01hy4rfqk8z4xv1c2v44cf6abc');
    expect(apiKey?.name).toBe('Personal Automation');
    expect(apiKey?.createdAt).toBeInstanceOf(Date);
    expect(apiKey?.deletedAt).toBeNull();
  });

  it('will parse the deleted at timestamp when the key has been revoked', async () => {
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc', // 1,
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk', // 4,
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: '2023-08-01T04:22:52.48118Z',
      },
    ]);

    const world = testRenderHook(useApiKeys, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.data).toHaveLength(1));

    const apiKey = world.result.current.data?.[0];
    expect(apiKey?.deletedAt).toBeInstanceOf(Date);
  });

  it('will return an empty array when there are no api keys', async () => {
    mockFetch.onGet('/api/keys').reply(200, []);

    const world = testRenderHook(useApiKeys, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.data).toBeDefined());
    await waitFor(() => expect(world.result.current.data).toHaveLength(0));
  });
});
