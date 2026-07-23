import { act } from 'react';
import { waitFor } from '@testing-library/react';

import useApiKeys from '@monetr/interface/hooks/useApiKeys';
import useRemoveApiKey from '@monetr/interface/hooks/useRemoveApiKey';
import type ApiKey from '@monetr/interface/models/ApiKey';
import { ID } from '@monetr/interface/models/ID';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

// The remove hook prunes the api key list out of the query cache when it succeeds, so render the list hook alongside it.
// That way we have a real cache entry to assert against instead of an empty one.
function useApiKeysAndRemove() {
  return {
    apiKeys: useApiKeys(),
    removeApiKey: useRemoveApiKey(),
  };
}

describe('remove api key', () => {
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

  function mockApiKeys() {
    mockFetch.onGet('/api/keys').reply(200, [
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc', // 1,
        name: 'Personal Automation',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk', // 4,
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
      {
        apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6xyz', // 2,
        name: 'CI Deploys',
        createdAt: '2023-07-02T04:22:52.48118Z',
        createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk', // 4,
        updatedAt: '2023-07-02T04:22:52.48118Z',
        deletedAt: null,
      },
    ]);
  }

  it('will remove an api key', async () => {
    mockApiKeys();
    mockFetch.onDelete('/api/keys/key_01hy4rfqk8z4xv1c2v44cf6abc').reply(200);

    const world = testRenderHook(useApiKeysAndRemove, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.apiKeys.data).toHaveLength(2));

    await act(async () => {
      await world.result.current.removeApiKey({
        apiKeyId: ID.from<ApiKey>('key_01hy4rfqk8z4xv1c2v44cf6abc'),
      });
    });

    // Make sure that we did make the API call. The history is keyed by method so typescript thinks the delete bucket
    // might be undefined, pull it into a local and assert its there before we poke at the first entry.
    const deleteHistory = mockFetch.history.delete;
    expect(deleteHistory).toHaveLength(1);
    expect(deleteHistory?.[0]).toMatchObject({ url: '/api/keys/key_01hy4rfqk8z4xv1c2v44cf6abc' });

    // The removed key should be pruned out of the cached list without us having to refetch it, the other key must
    // survive.
    await waitFor(() => expect(world.result.current.apiKeys.data).toHaveLength(1));
    expect(world.result.current.apiKeys.data?.[0]?.apiKeyId).toBe('key_01hy4rfqk8z4xv1c2v44cf6xyz');
  });

  it('will not send a body when there is no challenge', async () => {
    mockApiKeys();
    mockFetch.onDelete('/api/keys/key_01hy4rfqk8z4xv1c2v44cf6abc').reply(200);

    const world = testRenderHook(useApiKeysAndRemove, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.apiKeys.data).toHaveLength(2));

    await act(async () => {
      await world.result.current.removeApiKey({
        apiKeyId: ID.from<ApiKey>('key_01hy4rfqk8z4xv1c2v44cf6abc'),
      });
    });

    const deleteHistory = mockFetch.history.delete;
    expect(deleteHistory).toHaveLength(1);
    expect(deleteHistory?.[0]?.data).toBeUndefined();
  });

  it('will send the proof of work challenge when one is provided', async () => {
    mockApiKeys();
    mockFetch.onDelete('/api/keys/key_01hy4rfqk8z4xv1c2v44cf6abc').reply(200);

    const world = testRenderHook(useApiKeysAndRemove, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.apiKeys.data).toHaveLength(2));

    await act(async () => {
      await world.result.current.removeApiKey({
        apiKeyId: ID.from<ApiKey>('key_01hy4rfqk8z4xv1c2v44cf6abc'),
        challenge: 'abc123',
        nonce: 1234,
      });
    });

    // Grab the body that actually went over the wire so we can prove the challenge makes it to the server, and that the
    // api key id stays a path param instead of leaking into the body.
    const deleteHistory = mockFetch.history.delete;
    expect(deleteHistory).toHaveLength(1);
    const body = deleteHistory?.[0]?.data as Record<string, unknown>;
    expect(body.challenge).toBe('abc123');
    expect(body.nonce).toBe(1234);
    expect('apiKeyId' in body).toBe(false);
  });

  it('it will fail to remove an api key', async () => {
    mockApiKeys();
    mockFetch.onDelete('/api/keys/key_01hy4rfqk8z4xv1c2v44cf6abc').reply(400, {
      error: 'Invalid api key or something',
    });

    const world = testRenderHook(useApiKeysAndRemove, {
      initialRoute: '/',
    });
    await waitFor(() => expect(world.result.current.apiKeys.data).toHaveLength(2));

    await act(async () => {
      await expect(
        world.result.current.removeApiKey({
          apiKeyId: ID.from<ApiKey>('key_01hy4rfqk8z4xv1c2v44cf6abc'),
        }),
      ).rejects.toMatchObject({
        message: 'Request failed with status code 400',
        response: {
          data: {
            error: 'Invalid api key or something',
          },
        },
      });
    });

    // A failed removal must leave the cached list alone.
    expect(world.result.current.apiKeys.data).toHaveLength(2);
  });
});
