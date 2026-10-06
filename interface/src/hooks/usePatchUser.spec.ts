import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import type { Authentication } from '@monetr/interface/hooks/useAuthentication';
import { type PatchUserResponse, usePatchUser } from '@monetr/interface/hooks/usePatchUser';
import { ID } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';
import type User from '@monetr/interface/models/User';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

const userId = 'user_01hy4rbb1gjdek7h2xmgy5pnwk';
const firstLink = 'link_01hy4rcmadc01d2kzv7vynbxxx';
const secondLink = 'link_01hy4re7c1xc2v44cf6kx302jx';

function userJson(linkOrder: Array<string> | null) {
  return {
    userId,
    loginId: 'lgn_01hy4rbb1gjdek7h2xmgy5pnwk',
    login: {
      loginId: 'lgn_01hy4rbb1gjdek7h2xmgy5pnwk',
      email: 'someone@example.com',
      firstName: 'Alex',
      lastName: 'Rivera',
      totpEnabledAt: null,
    },
    accountId: 'acct_01hy4rbb1gjdek7h2xmgy5pnwk',
    account: {
      accountId: 'acct_01hy4rbb1gjdek7h2xmgy5pnwk',
      timezone: 'America/Chicago',
      locale: 'en_US',
      subscriptionActiveUntil: null,
      subscriptionStatus: null,
    },
    role: 'owner',
    linkOrder,
  };
}

function renderPatchUser() {
  const world = testRenderHook(
    () => ({
      patchUser: usePatchUser(),
      queryClient: useQueryClient(),
    }),
    { initialRoute: '/' },
  );

  // The optimistic update only touches /api/users/me when there is something cached there, which in the real app is
  // always the case since it's loaded on startup. Seed it with the original order.
  act(() => {
    world.result.current.queryClient.setQueryData(['GET', '/api/users/me'], {
      user: userJson([firstLink, secondLink]),
      isSetup: true,
      isActive: true,
    });
  });

  return world;
}

function cachedLinkOrder(world: ReturnType<typeof renderPatchUser>): Array<ID<Link>> | undefined {
  return world.result.current.queryClient.getQueryData<Partial<Authentication>>(['GET', '/api/users/me'])?.user
    ?.linkOrder;
}

describe('patch user', () => {
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

  it('will update the link order', async () => {
    mockFetch.onPatch(`/api/users/${userId}`).reply(200, userJson([secondLink, firstLink]));

    const world = renderPatchUser();

    let result!: PatchUserResponse;
    await act(async () => {
      result = await world.result.current.patchUser({
        userId: ID.from<User>(userId),
        linkOrder: [ID.from<Link>(secondLink), ID.from<Link>(firstLink)],
      });
    });

    expect(result.linkOrder).toEqual([secondLink, firstLink]);

    // The user ID goes in the path, it should not also end up in the body since the server rejects unknown fields.
    const patchHistory = mockFetch.history.patch;
    expect(patchHistory).toHaveLength(1);
    expect(patchHistory?.[0]?.url).toBe(`/api/users/${userId}`);
    expect(patchHistory?.[0]?.data).toEqual({ linkOrder: [secondLink, firstLink] });

    // The single user cache gets the response as is.
    const single = world.result.current.queryClient.getQueryData<PatchUserResponse>(['GET', `/api/users/${userId}`]);
    expect(single?.linkOrder).toEqual([secondLink, firstLink]);

    // And the current user keeps the new order.
    expect(cachedLinkOrder(world)).toEqual([secondLink, firstLink]);
  });

  it('will update the current user before the request finishes', async () => {
    mockFetch.onPatch(`/api/users/${userId}`).reply(200, userJson([secondLink, firstLink]));

    // Hold the request open so we can look at the cache while it's still in flight. This is the whole point of the
    // optimistic update, a reorder should show up right away instead of waiting on the server.
    const mockedFetch = globalThis.fetch;
    let release!: () => void;
    const gate = new Promise<void>(resolve => {
      release = resolve;
    });
    globalThis.fetch = async (...args) => {
      await gate;
      return mockedFetch(...args);
    };

    try {
      const world = renderPatchUser();

      let pending!: Promise<PatchUserResponse>;
      await act(async () => {
        pending = world.result.current.patchUser({
          userId: ID.from<User>(userId),
          linkOrder: [ID.from<Link>(secondLink), ID.from<Link>(firstLink)],
        });
        // Let onMutate finish, it awaits cancelQueries before it writes to the cache.
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(mockFetch.history.patch).toHaveLength(0);
      expect(cachedLinkOrder(world)).toEqual([secondLink, firstLink]);

      await act(async () => {
        release();
        await pending;
      });

      expect(cachedLinkOrder(world)).toEqual([secondLink, firstLink]);
    } finally {
      globalThis.fetch = mockedFetch;
    }
  });

  it('will put the current user back if the patch fails', async () => {
    mockFetch.onPatch(`/api/users/${userId}`).reply(400, {
      error: 'Invalid request',
    });

    const world = renderPatchUser();

    await act(async () => {
      await expect(
        world.result.current.patchUser({
          userId: ID.from<User>(userId),
          linkOrder: [ID.from<Link>(secondLink), ID.from<Link>(firstLink)],
        }),
      ).rejects.toMatchObject({
        message: 'Request failed with status code 400',
      });
    });

    expect(cachedLinkOrder(world)).toEqual([firstLink, secondLink]);
    // Nothing should have been cached for the single user since the patch never went through.
    expect(world.result.current.queryClient.getQueryData(['GET', `/api/users/${userId}`])).toBeUndefined();
  });

  it('will not create a current user cache if there was not one', async () => {
    mockFetch.onPatch(`/api/users/${userId}`).reply(200, userJson([secondLink, firstLink]));

    const world = testRenderHook(
      () => ({
        patchUser: usePatchUser(),
        queryClient: useQueryClient(),
      }),
      { initialRoute: '/' },
    );

    await act(async () => {
      await world.result.current.patchUser({
        userId: ID.from<User>(userId),
        linkOrder: [ID.from<Link>(secondLink), ID.from<Link>(firstLink)],
      });
    });

    expect(world.result.current.queryClient.getQueryData(['GET', '/api/users/me'])).toBeUndefined();
  });
});
