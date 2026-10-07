import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { useUpdateSpending } from '@monetr/interface/hooks/useUpdateSpending';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('update spending', () => {
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

  it('will leave the spending list alone when it was never loaded', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/spending/spnd_01hy4rkq0x3c6dtr9w1p2v5bns')
      .reply(200, {
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
        spendingType: 'expense',
        name: 'Netflix Premium',
        targetAmount: 2499,
        currentAmount: 0,
        usedAmount: 0,
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
        lastRecurrence: null,
        nextRecurrence: '2099-04-15T05:00:00Z',
        nextContributionAmount: 1250,
        isBehind: false,
        isPaused: false,
        autoCreateTransaction: false,
        createdAt: '2026-03-16T06:00:00Z',
      });

    const world = testRenderHook(
      () => ({
        updateSpending: useUpdateSpending(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/expenses',
      },
    );
    await act(async () => {
      await world.result.current.updateSpending({
        spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Netflix Premium',
        targetAmount: 2499,
      });
    });

    // Mapping over nothing would cache an empty list and hide every expense until something refetched it
    expect(
      world.result.current.queryClient.getQueryData([
        'GET',
        '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/spending',
      ]),
    ).toBeUndefined();
    expect(
      world.result.current.queryClient.getQueryData([
        'GET',
        '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/spending/spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
      ]),
    ).toBeDefined();
  });
});
