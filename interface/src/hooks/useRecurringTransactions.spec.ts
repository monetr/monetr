import { act, waitFor } from '@testing-library/react';

import { useRecurringTransactions } from '@monetr/interface/hooks/useRecurringTransactions';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('list recurring transactions', () => {
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

  it('will request the first page', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring?direction=debit&ended=false')
      .reply(200, [
        {
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
          spendingId: null,
          fundingScheduleId: null,
          window: 'monthly',
          ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
          first: '2026-01-15T06:00:00Z',
          last: '2026-03-15T05:00:00Z',
          next: '2026-04-15T05:00:00Z',
          ended: false,
          confidence: 0.9,
          direction: 'debit',
          amounts: {
            800: 3,
          },
          lastAmount: 800,
          autoMatched: false,
          autoAssign: false,
          createdAt: '2026-03-15T06:00:00Z',
          updatedAt: '2026-03-15T06:00:00Z',
        },
      ]);

    const world = testRenderHook(
      () =>
        useRecurringTransactions({
          direction: 'debit',
          ended: false,
        }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring',
      },
    );
    await waitFor(() => expect(world.result.current.data).toHaveLength(1));
    // Less than a full page means there isn't another one
    expect(world.result.current.hasNextPage).toBeFalsy();
  });

  it('will request the next page when the first one is full', async () => {
    // A full page is 25, the IDs just need to be different from each other
    mockFetch.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring?direction=debit&ended=false').reply(
      200,
      Array.from(
        {
          length: 25,
        },
        (_, index) => ({
          transactionRecurringId: `txrc_01hy4re7c1xc2v44cf6kx302${index.toString().padStart(2, '0')}`,
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
          spendingId: null,
          fundingScheduleId: null,
          window: 'monthly',
          ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
          first: '2026-01-15T06:00:00Z',
          last: '2026-03-15T05:00:00Z',
          next: '2026-04-15T05:00:00Z',
          ended: false,
          confidence: 0.9,
          direction: 'debit',
          amounts: {
            800: 3,
          },
          lastAmount: 800,
          autoMatched: false,
          autoAssign: false,
          createdAt: '2026-03-15T06:00:00Z',
          updatedAt: '2026-03-15T06:00:00Z',
        }),
      ),
    );
    mockFetch
      .onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring?direction=debit&ended=false&offset=25')
      .reply(200, [
        {
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx30225',
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
          spendingId: null,
          fundingScheduleId: null,
          window: 'monthly',
          ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
          first: '2026-01-15T06:00:00Z',
          last: '2026-03-15T05:00:00Z',
          next: '2026-04-15T05:00:00Z',
          ended: false,
          confidence: 0.9,
          direction: 'debit',
          amounts: {
            800: 3,
          },
          lastAmount: 800,
          autoMatched: false,
          autoAssign: false,
          createdAt: '2026-03-15T06:00:00Z',
          updatedAt: '2026-03-15T06:00:00Z',
        },
      ]);

    const world = testRenderHook(
      () =>
        useRecurringTransactions({
          direction: 'debit',
          ended: false,
        }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring',
      },
    );
    await waitFor(() => expect(world.result.current.data).toHaveLength(25));
    expect(world.result.current.hasNextPage).toBeTruthy();

    await act(async () => {
      await world.result.current.fetchNextPage();
    });
    await waitFor(() => expect(world.result.current.data).toHaveLength(26));
    expect(world.result.current.hasNextPage).toBeFalsy();
  });

  it('will not crash when the request fails to send', async () => {
    // Nothing is mocked so fetch will reject like it does when the network drops out

    const world = testRenderHook(
      () =>
        useRecurringTransactions({
          direction: 'debit',
          ended: false,
        }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring',
      },
    );
    // Only wait for the first failure, otherwise we are waiting on all of the retries
    await waitFor(() => expect(world.result.current.failureCount).toBe(1));
    expect(world.result.current.data).toBeUndefined();
  });
});
