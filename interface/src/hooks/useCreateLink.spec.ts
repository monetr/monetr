import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { useCreateLink } from '@monetr/interface/hooks/useCreateLink';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('create link', () => {
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
    mockFetch.onPost('/api/links').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Elliot Budget',
      description: null,
      createdAt: '2026-03-16T06:00:00Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-03-16T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderHook(
      () => ({
        createLink: useCreateLink(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/setup',
      },
    );
    await act(async () => {
      await world.result.current.createLink({
        institutionName: 'Elliot Budget',
      });
    });

    // Caching a list with just the new one in it would hide every other link until something refetched it
    expect(world.result.current.queryClient.getQueryData(['GET', '/api/links'])).toBeUndefined();
    expect(
      world.result.current.queryClient.getQueryData(['GET', '/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk']),
    ).toBeDefined();
  });
});
