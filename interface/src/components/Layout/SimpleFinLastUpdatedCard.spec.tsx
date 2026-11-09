import { waitFor } from '@testing-library/react';

import SimpleFinLastUpdatedCard from '@monetr/interface/components/Layout/SimpleFinLastUpdatedCard';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('simplefin last updated card', () => {
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

  it('will render relative timestamps', async () => {
    const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString();
    const twoHoursAgo = new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString();
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'simplefin',
      simpleFinLinkId: 'sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
      institutionName: 'SimpleFIN',
      simpleFinLink: {
        simpleFinLinkId: 'sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
        name: 'SimpleFIN',
        apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
        status: 'active',
        lastErrors: [],
        lastManualSync: null,
        lastSuccessfulUpdate: twoHoursAgo,
        lastAttemptedUpdate: fiveMinutesAgo,
        createdAt: '2026-03-15T06:00:00Z',
        createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        updatedAt: '2026-03-16T06:00:00Z',
        deletedAt: null,
      },
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });

    const world = testRenderer(<SimpleFinLastUpdatedCard linkId='link_01hy4rbb1gjdek7h2xmgy5pnwk' />);

    await waitFor(() => expect(world.getByText('Last Updated: about 2 hours ago')).toBeVisible());
    expect(world.getByText('Last Attempt: 5 minutes ago')).toBeVisible();
  });

  it('will show never when the link has not synced', async () => {
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'simplefin',
      simpleFinLinkId: 'sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
      institutionName: 'SimpleFIN',
      simpleFinLink: {
        simpleFinLinkId: 'sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
        name: 'SimpleFIN',
        apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
        status: 'active',
        lastErrors: null,
        lastManualSync: null,
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: null,
        createdAt: '2026-03-15T06:00:00Z',
        createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        updatedAt: '2026-03-16T06:00:00Z',
        deletedAt: null,
      },
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });

    const world = testRenderer(<SimpleFinLastUpdatedCard linkId='link_01hy4rbb1gjdek7h2xmgy5pnwk' />);

    await waitFor(() => expect(world.getByText('Last Updated: Never')).toBeVisible());
    expect(world.getByText('Last Attempt: Never')).toBeVisible();
  });

  it('will show never for updated when only attempts have been made', async () => {
    const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString();
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'simplefin',
      simpleFinLinkId: 'sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
      institutionName: 'SimpleFIN',
      simpleFinLink: {
        simpleFinLinkId: 'sfx_01hy4rbb1gjdek7h2xmgy5pnwk',
        name: 'SimpleFIN',
        apiUrl: 'https://beta-bridge.simplefin.org/simplefin',
        status: 'active',
        lastErrors: [],
        lastManualSync: null,
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: fiveMinutesAgo,
        createdAt: '2026-03-15T06:00:00Z',
        createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        updatedAt: '2026-03-16T06:00:00Z',
        deletedAt: null,
      },
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });

    const world = testRenderer(<SimpleFinLastUpdatedCard linkId='link_01hy4rbb1gjdek7h2xmgy5pnwk' />);

    await waitFor(() => expect(world.getByText('Last Updated: Never')).toBeVisible());
    expect(world.getByText('Last Attempt: 5 minutes ago')).toBeVisible();
  });

  it('will not render for a link without simplefin', async () => {
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      simpleFinLinkId: null,
      simpleFinLink: null,
      plaidLink: null,
      institutionName: 'Manual',
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });

    const world = testRenderer(<SimpleFinLastUpdatedCard linkId='link_01hy4rbb1gjdek7h2xmgy5pnwk' />);

    await waitFor(() =>
      expect(mockFetch.history.get?.map(item => item.url)).toContain('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk'),
    );
    expect(world.queryByText(/Last Updated/)).toBeNull();
  });
});
