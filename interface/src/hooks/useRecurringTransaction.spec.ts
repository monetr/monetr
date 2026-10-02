import { waitFor } from '@testing-library/react';

import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('read individual recurring transactions', () => {
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

  it('will request a single recurring transaction', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        window: 'monthly',
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
        first: '2026-01-15T06:00:00Z',
        last: '2026-03-15T05:00:00Z',
        next: '2026-04-15T05:00:00Z',
        ended: false,
        confidence: 0.9,
        direction: 'debit',
        amounts: { 800: 3 },
        lastAmount: 800,
        createdAt: '2026-03-15T06:00:00Z',
        updatedAt: '2026-03-15T06:00:00Z',
      });

    const world = testRenderHook(
      () => useRecurringTransaction(ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx')),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
      },
    );
    await waitFor(() => expect(world.result.current.isLoading).toBeTruthy());
    await waitFor(() => expect(world.result.current.data).toBeDefined());
    expect(world.result.current.data?.transactionRecurringId.toString()).toBe('txrc_01hy4re7c1xc2v44cf6kx302jx');
    expect(world.result.current.data?.direction).toBe('debit');
    expect(world.result.current.data?.window).toBe(TransactionRecurringWindow.Monthly);
    expect(world.result.current.data?.lastAmount).toBe(800);
    expect(world.result.current.data?.next).toBeInstanceOf(Date);
  });

  it('will not request anything without an id', async () => {
    const world = testRenderHook(() => useRecurringTransaction(null), {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
    });
    await waitFor(() => expect(world.result.current.fetchStatus).toBe('idle'));
    expect(world.result.current.data).toBeUndefined();
  });
});
