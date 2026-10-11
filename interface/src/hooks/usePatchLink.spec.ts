import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { usePatchLink } from '@monetr/interface/hooks/usePatchLink';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('patch link', () => {
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

  it('will leave the link list alone when it was never loaded', async () => {
    mockFetch.onPatch('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Elliot Budget',
      description: 'Joint account',
      createdAt: '2026-03-16T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-03-17T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderHook(
      () => ({
        patchLink: usePatchLink(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/settings/overview',
      },
    );
    await act(async () => {
      await world.result.current.patchLink({
        linkId: ID.from<Link>('link_01hy4rbb1gjdek7h2xmgy5pnwk'),
        description: 'Joint account',
      });
    });

    // Mapping over nothing would cache an empty list and hide every link until something refetched it
    expect(world.result.current.queryClient.getQueryData(['GET', '/api/links'])).toBeUndefined();
    expect(
      world.result.current.queryClient.getQueryData(['GET', '/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk']),
    ).toBeDefined();
  });

  it('will update the cached links with the new account order', async () => {
    mockFetch.onPatch('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Mercury',
      description: null,
      bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbyyy', 'bac_01hy4rcmadc01d2kzv7vynbxxx'],
      createdAt: '2026-03-16T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-03-17T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderHook(
      () => ({
        patchLink: usePatchLink(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
      },
    );

    // The sidebar and the details page both have the links loaded already.
    world.result.current.queryClient.setQueryData(
      ['GET', '/api/links'],
      [
        {
          linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
          linkType: 'manual',
          institutionName: 'Mercury',
          description: null,
          bankAccountOrder: null,
          createdAt: '2026-03-16T06:00:00Z',
          createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
          updatedAt: '2026-03-16T06:00:00Z',
          deletedAt: null,
        },
      ],
    );
    world.result.current.queryClient.setQueryData(['GET', '/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk'], {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Mercury',
      description: null,
      bankAccountOrder: null,
      createdAt: '2026-03-16T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-03-16T06:00:00Z',
      deletedAt: null,
    });

    await act(async () => {
      await world.result.current.patchLink({
        linkId: ID.from<Link>('link_01hy4rbb1gjdek7h2xmgy5pnwk'),
        bankAccountOrder: [
          ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbyyy'),
          ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        ],
      });
    });

    expect(mockFetch.history.patch?.[0]?.data).toMatchObject({
      bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbyyy', 'bac_01hy4rcmadc01d2kzv7vynbxxx'],
    });
    expect(world.result.current.queryClient.getQueryData(['GET', '/api/links'])).toMatchObject([
      {
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
        bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbyyy', 'bac_01hy4rcmadc01d2kzv7vynbxxx'],
      },
    ]);
    expect(
      world.result.current.queryClient.getQueryData(['GET', '/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk']),
    ).toMatchObject({
      bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbyyy', 'bac_01hy4rcmadc01d2kzv7vynbxxx'],
    });
  });

  it('will put the account order back when saving it fails', async () => {
    mockFetch.onPatch('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(400, {
      error: 'Invalid request',
    });

    const world = testRenderHook(
      () => ({
        patchLink: usePatchLink(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
      },
    );

    // The user already had an order saved from before.
    world.result.current.queryClient.setQueryData(
      ['GET', '/api/links'],
      [
        {
          linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
          linkType: 'manual',
          institutionName: 'Mercury',
          description: null,
          bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbxxx', 'bac_01hy4rcmadc01d2kzv7vynbyyy'],
          createdAt: '2026-03-16T06:00:00Z',
          createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
          updatedAt: '2026-03-16T06:00:00Z',
          deletedAt: null,
        },
      ],
    );
    world.result.current.queryClient.setQueryData(['GET', '/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk'], {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Mercury',
      description: null,
      bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbxxx', 'bac_01hy4rcmadc01d2kzv7vynbyyy'],
      createdAt: '2026-03-16T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-03-16T06:00:00Z',
      deletedAt: null,
    });

    // Drag the savings account to the top, but the server rejects it.
    await act(async () => {
      await expect(
        world.result.current.patchLink({
          linkId: ID.from<Link>('link_01hy4rbb1gjdek7h2xmgy5pnwk'),
          bankAccountOrder: [
            ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbyyy'),
            ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
          ],
        }),
      ).rejects.toBeDefined();
    });

    // Both caches should be back to the order from before the drag.
    expect(world.result.current.queryClient.getQueryData(['GET', '/api/links'])).toMatchObject([
      {
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
        bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbxxx', 'bac_01hy4rcmadc01d2kzv7vynbyyy'],
      },
    ]);
    expect(
      world.result.current.queryClient.getQueryData(['GET', '/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk']),
    ).toMatchObject({
      bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbxxx', 'bac_01hy4rcmadc01d2kzv7vynbyyy'],
    });
  });
});
