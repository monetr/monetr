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

const bankAccountId = 'bac_01hy4rcmadc01d2kzv7vynbxxx';
const recurringId = 'txrc_01hy4re7c1xc2v44cf6kx302jx';
const spendingId = 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns';
const fundingScheduleId = 'fund_01hy4re7c1xc2v44cf6kx302jx';
const recurringKey = ['GET', `/api/bank_accounts/${bankAccountId}/recurring/${recurringId}`];

function recurringJson(linked: boolean) {
  return {
    transactionRecurringId: recurringId,
    bankAccountId,
    transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
    spendingId: linked ? spendingId : null,
    fundingScheduleId: null,
    window: 'monthly',
    ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
    first: '2026-01-15T06:00:00Z',
    last: '2026-03-15T05:00:00Z',
    next: '2026-04-15T05:00:00Z',
    ended: false,
    confidence: 0.9,
    direction: 'debit',
    amounts: { 1549: 3 },
    lastAmount: 1549,
    createdAt: '2026-01-15T06:00:00Z',
    updatedAt: '2026-03-16T06:00:00Z',
  };
}

function renderPatchTransactionRecurring(cached: ReturnType<typeof recurringJson>) {
  const world = testRenderHook(
    () => ({
      patchTransactionRecurring: usePatchTransactionRecurring(),
      queryClient: useQueryClient(),
    }),
    { initialRoute: `/bank/${bankAccountId}/transactions` },
  );

  // In the app the recurring transaction is already cached from the details page by the time it gets patched.
  act(() => {
    world.result.current.queryClient.setQueryData(recurringKey, cached);
  });

  return world;
}

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
    mockFetch.onPatch(`/api/bank_accounts/${bankAccountId}/recurring/${recurringId}`).reply(200, recurringJson(true));

    const world = renderPatchTransactionRecurring(recurringJson(false));

    let result!: PatchTransactionRecurringResponse;
    await act(async () => {
      result = await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>(recurringId),
        bankAccountId: ID.from<BankAccount>(bankAccountId),
        spendingId: ID.from<Spending>(spendingId),
      });
    });

    expect(result.spendingId).toBe(spendingId);

    // Both IDs go in the path, they should not also end up in the body since the server rejects unknown fields.
    const patchHistory = mockFetch.history.patch;
    expect(patchHistory).toHaveLength(1);
    expect(patchHistory?.[0]?.url).toBe(`/api/bank_accounts/${bankAccountId}/recurring/${recurringId}`);
    expect(patchHistory?.[0]?.data).toEqual({ spendingId });

    // The cached recurring transaction gets replaced with the response, linked expense and all.
    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>(recurringKey);
    expect(cached?.spendingId).toBe(spendingId);
  });

  it('will refresh the recurring list', async () => {
    mockFetch.onPatch(`/api/bank_accounts/${bankAccountId}/recurring/${recurringId}`).reply(200, recurringJson(true));

    const world = renderPatchTransactionRecurring(recurringJson(false));
    const listKey = ['GET', `/api/bank_accounts/${bankAccountId}/recurring`];
    act(() => {
      world.result.current.queryClient.setQueryData(listKey, { pages: [[recurringJson(false)]], pageParams: [0] });
    });

    await act(async () => {
      await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>(recurringId),
        bankAccountId: ID.from<BankAccount>(bankAccountId),
        spendingId: ID.from<Spending>(spendingId),
      });
    });

    // The list items have the cluster on them which isn't in the response, so the list just gets marked stale.
    expect(world.result.current.queryClient.getQueryState(listKey)?.isInvalidated).toBeTruthy();
  });

  it('will link a funding schedule', async () => {
    mockFetch.onPatch(`/api/bank_accounts/${bankAccountId}/recurring/${recurringId}`).reply(200, {
      ...recurringJson(false),
      direction: 'credit',
      fundingScheduleId,
    });

    const world = renderPatchTransactionRecurring({ ...recurringJson(false), direction: 'credit' });

    await act(async () => {
      await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>(recurringId),
        bankAccountId: ID.from<BankAccount>(bankAccountId),
        fundingScheduleId: ID.from<FundingSchedule>(fundingScheduleId),
      });
    });

    expect(mockFetch.history.patch?.[0]?.data).toEqual({ fundingScheduleId });
    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>(recurringKey);
    expect(cached?.fundingScheduleId).toBe(fundingScheduleId);
  });

  it('will send a null to unlink', async () => {
    mockFetch.onPatch(`/api/bank_accounts/${bankAccountId}/recurring/${recurringId}`).reply(200, recurringJson(false));

    const world = renderPatchTransactionRecurring(recurringJson(true));

    await act(async () => {
      await world.result.current.patchTransactionRecurring({
        transactionRecurringId: ID.from<TransactionRecurring>(recurringId),
        bankAccountId: ID.from<BankAccount>(bankAccountId),
        spendingId: null,
      });
    });

    // The null has to make it to the server, thats how it knows to clear the link instead of leaving it alone.
    expect(mockFetch.history.patch?.[0]?.data).toEqual({ spendingId: null });
    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>(recurringKey);
    expect(cached?.spendingId).toBeNull();
  });

  it('will leave the cache alone if the patch fails', async () => {
    mockFetch.onPatch(`/api/bank_accounts/${bankAccountId}/recurring/${recurringId}`).reply(400, {
      error: 'failed to update recurring transaction: a similar object already exists',
    });

    const world = renderPatchTransactionRecurring(recurringJson(false));

    await act(async () => {
      await expect(
        world.result.current.patchTransactionRecurring({
          transactionRecurringId: ID.from<TransactionRecurring>(recurringId),
          bankAccountId: ID.from<BankAccount>(bankAccountId),
          spendingId: ID.from<Spending>(spendingId),
        }),
      ).rejects.toMatchObject({
        message: 'Request failed with status code 400',
      });
    });

    const cached = world.result.current.queryClient.getQueryData<PatchTransactionRecurringResponse>(recurringKey);
    expect(cached?.spendingId).toBeNull();
  });
});
