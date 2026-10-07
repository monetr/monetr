import { act, waitFor } from '@testing-library/react';

import { useTransactionsForRecurring } from '@monetr/interface/hooks/useTransactionsForRecurring';
import { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('transactions for a recurring transaction', () => {
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

  it('will only request as many as it was asked for', async () => {
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=2',
      )
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v2',
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          amount: 800,
          date: '2026-03-15T05:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM 877-448-4820 CA',
          isPending: false,
          createdAt: '2026-03-15T06:00:00Z',
        },
      ]);

    const world = testRenderHook(
      () => useTransactionsForRecurring(ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'), 2),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring',
      },
    );
    await waitFor(() => expect(world.result.current.data).toHaveLength(1));
    // Less than a full page means there isn't another one
    expect(world.result.current.hasNextPage).toBeFalsy();
  });

  it('will request the next page when the first one is full', async () => {
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=2',
      )
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v2',
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          amount: 800,
          date: '2026-03-15T05:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM 877-448-4820 CA',
          isPending: false,
          createdAt: '2026-03-15T06:00:00Z',
        },
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v1',
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          amount: 800,
          date: '2026-02-15T06:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM 877-448-4820 CA',
          isPending: false,
          createdAt: '2026-02-15T07:00:00Z',
        },
      ]);
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=2&offset=2',
      )
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v0',
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          amount: 800,
          date: '2026-01-15T06:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM 877-448-4820 CA',
          isPending: false,
          createdAt: '2026-01-15T07:00:00Z',
        },
      ]);

    const world = testRenderHook(
      () => useTransactionsForRecurring(ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'), 2),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring',
      },
    );
    await waitFor(() => expect(world.result.current.data).toHaveLength(2));
    expect(world.result.current.hasNextPage).toBeTruthy();

    await act(async () => {
      await world.result.current.fetchNextPage();
    });
    await waitFor(() => expect(world.result.current.data).toHaveLength(3));
    expect(world.result.current.hasNextPage).toBeFalsy();
  });

  it('will not request anything without a recurring transaction', async () => {
    const world = testRenderHook(() => useTransactionsForRecurring(undefined, 10), {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring',
    });

    // This is how the details page keeps the tab that isn't showing from loading anything
    await waitFor(() => expect(world.result.current.fetchStatus).toBe('idle'));
    expect(mockFetch.history.get).toHaveLength(0);
  });
});
