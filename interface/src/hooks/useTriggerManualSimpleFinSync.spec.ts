import { act } from 'react';

import { useTriggerManualSimpleFinSync } from '@monetr/interface/hooks/useTriggerManualSimpleFinSync';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('useTriggerManualSimpleFinSync', () => {
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

  it('will trigger a manual sync', async () => {
    mockFetch.onPost('/api/simplefin/link/sync').reply(202);

    const world = testRenderHook(useTriggerManualSimpleFinSync, { initialRoute: '/' });

    expect(mockFetch.history.post).toHaveLength(0);

    await act(() => {
      return world.result.current('link_01hy4rbb1gjdek7h2xmgy5pnwk');
    });

    const postHistory = mockFetch.history.post;
    expect(postHistory).toHaveLength(1);
    expect(postHistory?.[0]).toMatchObject({
      url: '/api/simplefin/link/sync',
      data: {
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      },
    });
  });

  it('will not throw when synced too recently', async () => {
    mockFetch.onPost('/api/simplefin/link/sync').reply(425, {
      error: 'Link has been manually synced too recently',
    });

    const world = testRenderHook(useTriggerManualSimpleFinSync, { initialRoute: '/' });

    // The error should be shown as a snackbar, not bubble up to the menu item.
    await act(() => {
      return world.result.current('link_01hy4rbb1gjdek7h2xmgy5pnwk');
    });

    expect(mockFetch.history.post).toHaveLength(1);
  });
});
