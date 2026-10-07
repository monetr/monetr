import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { usePatchLink } from '@monetr/interface/hooks/usePatchLink';
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
});
