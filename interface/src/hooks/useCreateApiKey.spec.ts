import { act } from 'react';

import useCreateApiKey, { type CreateApiKeyResponse } from '@monetr/interface/hooks/useCreateApiKey';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('create api key', () => {
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

  it('will create an api key', async () => {
    mockFetch.onPost('/api/keys').reply(200, {
      apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc', // 1,
      name: 'Personal Automation',
      createdAt: '2023-07-02T04:22:52.48118Z',
      createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk', // 4,
      updatedAt: '2023-07-02T04:22:52.48118Z',
      deletedAt: null,
      secret: 'monetr_secret_aebagbafaydqqcikbmga2dqpcaireeyuculbogazdinryhi6d4qa',
    });

    const world = testRenderHook(useCreateApiKey, {
      initialRoute: '/',
    });
    let result!: CreateApiKeyResponse;
    await act(async () => {
      result = await world.result.current({
        name: 'Personal Automation',
      });
    });
    expect(result).toBeDefined();
    expect(result.apiKeyId).toBe('key_01hy4rfqk8z4xv1c2v44cf6abc');
    expect(result.name).toBe('Personal Automation');
    // The secret is only ever returned on create, it is not part of the ApiKey model itself so make sure the hook grafts
    // it onto the response instead of dropping it on the floor.
    expect(result.secret).toBe('monetr_secret_aebagbafaydqqcikbmga2dqpcaireeyuculbogazdinryhi6d4qa');
    // The rest of the payload should still be hydrated the way the model would do it.
    expect(result.createdAt).toBeInstanceOf(Date);
    expect(result.deletedAt).toBeNull();
  });

  it('will send the proof of work challenge when one is provided', async () => {
    mockFetch.onPost('/api/keys').reply(200, {
      apiKeyId: 'key_01hy4rfqk8z4xv1c2v44cf6abc', // 1,
      name: 'Personal Automation',
      createdAt: '2023-07-02T04:22:52.48118Z',
      createdBy: 'user_01hy4rbb1gjdek7h2xmgy5pnwk', // 4,
      updatedAt: '2023-07-02T04:22:52.48118Z',
      deletedAt: null,
      secret: 'monetr_secret_aebagbafaydqqcikbmga2dqpcaireeyuculbogazdinryhi6d4qa',
    });

    const world = testRenderHook(useCreateApiKey, {
      initialRoute: '/',
    });
    await act(async () => {
      await world.result.current({
        name: 'Personal Automation',
        challenge: 'abc123',
        nonce: 1234,
      });
    });

    // Grab the body that actually went over the wire so we can prove the challenge makes it to the server. The history
    // is keyed by method so typescript thinks the post bucket might be undefined, pull it into a local and assert its
    // there before we poke at the first entry.
    const postHistory = mockFetch.history.post;
    expect(postHistory).toHaveLength(1);
    const body = postHistory?.[0]?.data as Record<string, unknown>;
    expect(body.name).toBe('Personal Automation');
    expect(body.challenge).toBe('abc123');
    expect(body.nonce).toBe(1234);
  });

  it('it will fail to create an api key', async () => {
    mockFetch.onPost('/api/keys').reply(400, {
      error: 'Invalid api key or something',
    });

    const world = testRenderHook(useCreateApiKey, {
      initialRoute: '/',
    });
    await act(async () => {
      await expect(
        world.result.current({
          name: 'Personal Automation',
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
  });
});
