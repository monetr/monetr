import { act } from 'react';

import { useTriggerManualLunchFlowSync } from '@monetr/interface/hooks/useTriggerManualLunchFlowSync';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('useTriggerManualLunchFlowSync', () => {
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
    mockFetch.onPost('/api/lunch_flow/link/sync').reply(202);

    const world = testRenderHook(useTriggerManualLunchFlowSync, { initialRoute: '/' });

    expect(mockFetch.history.post).toHaveLength(0);

    await act(() => {
      return world.result.current('link_01hy4rbb1gjdek7h2xmgy5pnwk');
    });

    const postHistory = mockFetch.history.post;
    expect(postHistory).toHaveLength(1);
    expect(postHistory?.[0]).toMatchObject({
      url: '/api/lunch_flow/link/sync',
      data: {
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      },
    });
  });

  it('will not throw when synced too recently', async () => {
    mockFetch.onPost('/api/lunch_flow/link/sync').reply(425, {
      error: 'Link has been manually synced too recently',
    });

    const world = testRenderHook(useTriggerManualLunchFlowSync, { initialRoute: '/' });

    // The error should be shown as a snackbar, not bubble up to the menu item.
    await act(() => {
      return world.result.current('link_01hy4rbb1gjdek7h2xmgy5pnwk');
    });

    expect(mockFetch.history.post).toHaveLength(1);
  });
});
