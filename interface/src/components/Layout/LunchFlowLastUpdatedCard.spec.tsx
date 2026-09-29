import { waitFor } from '@testing-library/react';

import LunchFlowLastUpdatedCard from '@monetr/interface/components/Layout/LunchFlowLastUpdatedCard';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

const linkId = 'link_01lunchflow';

function lunchFlowLinkResponse(lunchFlowLink: {
  lastSuccessfulUpdate: string | null;
  lastAttemptedUpdate: string | null;
}) {
  return {
    linkId,
    linkType: 'lunch_flow',
    lunchFlowLinkId: 'lfx_01lunchflow',
    institutionName: 'Lunch Flow',
    lunchFlowLink: {
      lunchFlowLinkId: 'lfx_01lunchflow',
      name: 'Lunch Flow',
      apiUrl: 'https://www.lunchflow.app/api/v1',
      status: 'active',
      lastManualSync: null,
      updatedAt: '2025-01-01T00:00:00Z',
      createdAt: '2025-01-01T00:00:00Z',
      deletedAt: null,
      createdBy: 'user_01lunchflow',
      ...lunchFlowLink,
    },
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
  };
}

describe('lunch flow last updated card', () => {
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
    mockFetch.onGet(`/api/links/${linkId}`).reply(
      200,
      lunchFlowLinkResponse({
        lastSuccessfulUpdate: twoHoursAgo,
        lastAttemptedUpdate: fiveMinutesAgo,
      }),
    );

    const world = testRenderer(<LunchFlowLastUpdatedCard linkId={linkId} />);

    await waitFor(() => expect(world.getByText('Last Updated: about 2 hours ago')).toBeVisible());
    expect(world.getByText('Last Attempt: 5 minutes ago')).toBeVisible();
  });

  it('will show never when the link has not synced', async () => {
    mockFetch.onGet(`/api/links/${linkId}`).reply(
      200,
      lunchFlowLinkResponse({
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: null,
      }),
    );

    const world = testRenderer(<LunchFlowLastUpdatedCard linkId={linkId} />);

    await waitFor(() => expect(world.getByText('Last Updated: Never')).toBeVisible());
    expect(world.getByText('Last Attempt: Never')).toBeVisible();
  });

  it('will show never for updated when only attempts have been made', async () => {
    const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString();
    mockFetch.onGet(`/api/links/${linkId}`).reply(
      200,
      lunchFlowLinkResponse({
        lastSuccessfulUpdate: null,
        lastAttemptedUpdate: fiveMinutesAgo,
      }),
    );

    const world = testRenderer(<LunchFlowLastUpdatedCard linkId={linkId} />);

    await waitFor(() => expect(world.getByText('Last Updated: Never')).toBeVisible());
    expect(world.getByText('Last Attempt: 5 minutes ago')).toBeVisible();
  });

  it('will not render for a link without lunch flow', async () => {
    mockFetch.onGet(`/api/links/${linkId}`).reply(200, {
      linkId,
      linkType: 'manual',
      lunchFlowLinkId: null,
      lunchFlowLink: null,
      plaidLink: null,
      institutionName: 'Manual',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    });

    const world = testRenderer(<LunchFlowLastUpdatedCard linkId={linkId} />);

    await waitFor(() => expect(mockFetch.history.get?.map(item => item.url)).toContain(`/api/links/${linkId}`));
    expect(world.queryByText(/Last Updated/)).toBeNull();
  });
});
