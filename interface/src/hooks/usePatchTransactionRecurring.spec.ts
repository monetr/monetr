import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import {
  type PatchTransactionRecurringResponse,
  usePatchTransactionRecurring,
} from '@monetr/interface/hooks/usePatchTransactionRecurring';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('patch transaction recurring', () => {
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

  it('will link an expense', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
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
          1549: 3,
        },
        lastAmount: 1549,
        createdAt: '2026-01-15T06:00:00Z',
        updatedAt: '2026-03-16T06:00:00Z',
      });

    const world = testRenderHook(
      () => ({
        patchTransactionRecurring: usePatchTransactionRecurring(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
      },
    );
    // In the app the recurring transaction is already cached from the details page by the time it gets patched
    act(() => {
      world.result.current.queryClient.setQueryData(
        ['GET', '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx'],
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
            1549: 3,
          },
          lastAmount: 1549,
          createdAt: '2026-01-15T06:00:00Z',
          updatedAt: '2026-03-16T06:00:00Z',
        },
      );
    });

    let result!: PatchTransactionRecurringResponse;
    await act(async () => {
      result = await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
      });
    });

    expect(result.spendingId).toBe('spnd_01hy4rkq0x3c6dtr9w1p2v5bns');

    // Both IDs go in the path, they should not also end up in the body since the server rejects unknown fields.
    const patchHistory = mockFetch.history.patch;
    expect(patchHistory).toHaveLength(1);
    expect(patchHistory?.[0]?.url).toBe(
      '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx',
    );
    expect(patchHistory?.[0]?.data).toEqual({
      spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
    });

    // The cached recurring transaction gets replaced with the response, linked expense and all.
    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>([
      'GET',
      '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx',
    ]);
    expect(cached?.spendingId).toBe('spnd_01hy4rkq0x3c6dtr9w1p2v5bns');
  });

  it('will refresh the recurring list', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
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
          1549: 3,
        },
        lastAmount: 1549,
        createdAt: '2026-01-15T06:00:00Z',
        updatedAt: '2026-03-16T06:00:00Z',
      });

    const world = testRenderHook(
      () => ({
        patchTransactionRecurring: usePatchTransactionRecurring(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
      },
    );
    // In the app the recurring transaction is already cached from the details page by the time it gets patched
    act(() => {
      world.result.current.queryClient.setQueryData(
        ['GET', '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx'],
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
            1549: 3,
          },
          lastAmount: 1549,
          createdAt: '2026-01-15T06:00:00Z',
          updatedAt: '2026-03-16T06:00:00Z',
        },
      );
    });
    act(() => {
      world.result.current.queryClient.setQueryData(
        ['GET', '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring'],
        {
          pages: [
            [
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
                  1549: 3,
                },
                lastAmount: 1549,
                createdAt: '2026-01-15T06:00:00Z',
                updatedAt: '2026-03-16T06:00:00Z',
              },
            ],
          ],
          pageParams: [0],
        },
      );
    });

    await act(async () => {
      await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
      });
    });

    // The list items have the cluster on them which isn't in the response, so the list just gets marked stale.
    expect(
      world.result.current.queryClient.getQueryState([
        'GET',
        '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring',
      ])?.isInvalidated,
    ).toBeTruthy();
  });

  it('will link a funding schedule', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: null,
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
        window: 'monthly',
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
        first: '2026-01-15T06:00:00Z',
        last: '2026-03-15T05:00:00Z',
        next: '2026-04-15T05:00:00Z',
        ended: false,
        confidence: 0.9,
        direction: 'credit',
        amounts: {
          1549: 3,
        },
        lastAmount: 1549,
        createdAt: '2026-01-15T06:00:00Z',
        updatedAt: '2026-03-16T06:00:00Z',
      });

    const world = testRenderHook(
      () => ({
        patchTransactionRecurring: usePatchTransactionRecurring(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
      },
    );
    // In the app the recurring transaction is already cached from the details page by the time it gets patched
    act(() => {
      world.result.current.queryClient.setQueryData(
        ['GET', '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx'],
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
          direction: 'credit',
          amounts: {
            1549: 3,
          },
          lastAmount: 1549,
          createdAt: '2026-01-15T06:00:00Z',
          updatedAt: '2026-03-16T06:00:00Z',
        },
      );
    });

    await act(async () => {
      await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        fundingScheduleId: ID.from<FundingSchedule>('fund_01hy4re7c1xc2v44cf6kx302jx'),
      });
    });

    expect(mockFetch.history.patch?.[0]?.data).toEqual({
      fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
    });
    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>([
      'GET',
      '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx',
    ]);
    expect(cached?.fundingScheduleId).toBe('fund_01hy4re7c1xc2v44cf6kx302jx');
  });

  it('will send a null to unlink', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
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
          1549: 3,
        },
        lastAmount: 1549,
        createdAt: '2026-01-15T06:00:00Z',
        updatedAt: '2026-03-16T06:00:00Z',
      });

    const world = testRenderHook(
      () => ({
        patchTransactionRecurring: usePatchTransactionRecurring(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
      },
    );
    // In the app the recurring transaction is already cached from the details page by the time it gets patched
    act(() => {
      world.result.current.queryClient.setQueryData(
        ['GET', '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx'],
        {
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
          spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
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
            1549: 3,
          },
          lastAmount: 1549,
          createdAt: '2026-01-15T06:00:00Z',
          updatedAt: '2026-03-16T06:00:00Z',
        },
      );
    });

    await act(async () => {
      await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        spendingId: null,
      });
    });

    // The null has to make it to the server, that's how it knows to clear the link instead of leaving it alone.
    expect(mockFetch.history.patch?.[0]?.data).toEqual({
      spendingId: null,
    });
    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>([
      'GET',
      '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx',
    ]);
    expect(cached?.spendingId).toBeNull();
  });

  it('will leave the cache alone if the patch fails', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(400, {
        error: 'failed to update recurring transaction: a similar object already exists',
      });

    const world = testRenderHook(
      () => ({
        patchTransactionRecurring: usePatchTransactionRecurring(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
      },
    );
    // In the app the recurring transaction is already cached from the details page by the time it gets patched
    act(() => {
      world.result.current.queryClient.setQueryData(
        ['GET', '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx'],
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
            1549: 3,
          },
          lastAmount: 1549,
          createdAt: '2026-01-15T06:00:00Z',
          updatedAt: '2026-03-16T06:00:00Z',
        },
      );
    });

    await act(async () => {
      await expect(
        world.result.current.patchTransactionRecurring({
          transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
          bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
          spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
        }),
      ).rejects.toMatchObject({
        message: 'Request failed with status code 400',
      });
    });

    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>([
      'GET',
      '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx',
    ]);
    expect(cached?.spendingId).toBeNull();
  });
});
