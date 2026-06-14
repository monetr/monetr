import { act } from 'react';

import {
  type PatchFundingScheduleResponse,
  usePatchFundingSchedule,
} from '@monetr/interface/hooks/usePatchFundingSchedule';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import Spending from '@monetr/interface/models/Spending';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';
import parseDate from '@monetr/interface/util/parseDate';

describe('patch funding schedule', () => {
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

  it('will update a funding schedule', async () => {
    mockFetch.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
      availableBalance: 48635,
      currentBalance: 48635,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      officialName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      lastUpdated: '2023-07-02T04:22:52.48118Z',
    });
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        fundingSchedule: {
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
          dateStarted: '2023-02-28T06:00:00Z',
          description: '15th and last day of every month',
          estimatedDeposit: null,
          excludeWeekends: true,
          fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx', // 3,
          lastRecurrence: '2023-09-29T05:00:00Z',
          name: "Elliot's Contribution",
          nextRecurrence: '2023-10-13T05:00:00Z',
          nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
          ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          waitForDeposit: false,
        },
        spending: [],
      });

    const world = testRenderHook(usePatchFundingSchedule, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding',
    });
    let result!: PatchFundingScheduleResponse;
    await act(async () => {
      result = await world.result.current({
        fundingScheduleId: ID.from<FundingSchedule>('fund_01hy4re7c1xc2v44cf6kx302jx'), // 3,
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        description: 'something',
        name: "Elliot's Contribution",
        nextRecurrence: parseDate('2023-07-31T05:00:00Z') ?? undefined,
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        estimatedDeposit: null,
        excludeWeekends: true,
      });
    });
    expect(result).toBeDefined();
    expect(result.fundingSchedule.fundingScheduleId).toBe('fund_01hy4re7c1xc2v44cf6kx302jx');
    expect(result.spending.length).toBe(0);
  });

  it('it will fail to update a funding schedule', async () => {
    mockFetch.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
      availableBalance: 48635,
      currentBalance: 48635,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      officialName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      lastUpdated: '2023-07-02T04:22:52.48118Z',
    });
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx')
      .reply(400, {
        error: 'Invalid request',
      });

    const world = testRenderHook(usePatchFundingSchedule, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding',
    });
    await act(async () => {
      await expect(
        world.result.current({
          fundingScheduleId: ID.from<FundingSchedule>('fund_01hy4re7c1xc2v44cf6kx302jx'), // 3,
          bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
          description: 'something',
          name: "Elliot's Contribution",
          nextRecurrence: parseDate('2023-07-31T05:00:00Z') ?? undefined,
          ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
          estimatedDeposit: null,
          excludeWeekends: true,
        }),
      ).rejects.toMatchObject({
        message: 'Request failed with status code 400',
        response: {
          data: {
            error: 'Invalid request',
          },
        },
      });
    });
  });

  it('will drop undefined fields from the request body but keep nulls', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        fundingSchedule: {
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          description: null,
          estimatedDeposit: null,
          excludeWeekends: true,
          fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
          name: "Elliot's Contribution",
          nextRecurrence: '2023-10-13T05:00:00Z',
          ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        },
        spending: [],
      });

    const world = testRenderHook(usePatchFundingSchedule, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding',
    });
    await act(async () => {
      await world.result.current({
        fundingScheduleId: ID.from<FundingSchedule>('fund_01hy4re7c1xc2v44cf6kx302jx'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        name: "Elliot's Contribution",
        // description is left undefined on purpose, we did not touch it so it should not be sent at all.
        description: undefined,
        // estimatedDeposit is explicitly null, the user wants to clear it so it MUST survive into the request body.
        estimatedDeposit: null,
      });
    });

    // Grab the body that actually went over the wire so we can prove the undefined fields never get sent. The history
    // is keyed by method so typescript thinks the patch bucket might be undefined, pull it into a local and assert its
    // there before we poke at the first entry.
    const patchHistory = mockFetch.history.patch;
    expect(patchHistory).toHaveLength(1);
    const body = patchHistory?.[0]?.data as Record<string, unknown>;
    expect('description' in body).toBe(false);
    expect('estimatedDeposit' in body).toBe(true);
    expect(body.estimatedDeposit).toBeNull();
    expect(body.name).toBe("Elliot's Contribution");
    // fundingScheduleId and bankAccountId are path params, the hook destructures them out so they should never end up
    // in the patch body.
    expect('fundingScheduleId' in body).toBe(false);
    expect('bankAccountId' in body).toBe(false);
  });

  it('will hydrate the returned spending into Spending models', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules/fund_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        fundingSchedule: {
          bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
          description: '15th and last day of every month',
          estimatedDeposit: null,
          excludeWeekends: false,
          fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
          name: "Elliot's Contribution",
          nextRecurrence: '2023-10-13T05:00:00Z',
          ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        },
        spending: [
          {
            spendingId: 'spnd_01hy4rfqk8z4xv1c2v44cf6abc',
            bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
            fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
            name: 'Some Monthly Expense',
            description: null,
            targetAmount: 1000,
            currentAmount: 0,
            ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
            nextRecurrence: '2023-11-01T05:00:00Z',
            nextContributionAmount: 500,
            isPaused: false,
            autoCreateTransaction: false,
          },
        ],
      });

    const world = testRenderHook(usePatchFundingSchedule, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/funding',
    });
    let result!: PatchFundingScheduleResponse;
    await act(async () => {
      result = await world.result.current({
        fundingScheduleId: ID.from<FundingSchedule>('fund_01hy4re7c1xc2v44cf6kx302jx'),
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        excludeWeekends: false,
      });
    });

    expect(result.spending.length).toBe(1);
    // The hook should give us a real Spending instance back, not just the raw json, otherwise the getters everything
    // downstream relies on wont exist.
    const spending = result.spending[0];
    expect(spending).toBeInstanceOf(Spending);
    expect(spending?.spendingId).toBe('spnd_01hy4rfqk8z4xv1c2v44cf6abc');
  });
});
