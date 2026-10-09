import { waitFor } from '@testing-library/react';

import RecurringSummaryCard from '@monetr/interface/components/recurring/RecurringSummaryCard';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('recurring summary card', () => {
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

  it('will show the next couple of dates after the next one', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/similar/tcl_01hy4rf0p7mz9w2q3c4v5b6n7m')
      .reply(200, {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Github',
        originalMemo: 'GITHUB.COM 877-448-4820 CA',
        members: [],
        createdAt: '2026-01-15T06:00:00Z',
      });
    apiSampleResponses(mockFetch);

    // Next is midnight in Chicago during daylight savings, but the rule lands on 1am since it started in the winter.
    // Looking for the one after next from the same day used to just give back next again.
    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      spendingId: null,
      fundingScheduleId: null,
      fundingSchedule: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2099-04-15T05:00:00Z',
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
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSummaryCard onMarkNotRecurring={() => {}} recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() =>
      expect(world.getByTestId('recurring-next-later')).toHaveTextContent('Then May 15, 2099 and Jun 15, 2099'),
    );
  });

  it('will say the next charge is covered when the expense is on track', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/similar/tcl_01hy4rf0p7mz9w2q3c4v5b6n7m')
      .reply(200, {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Github',
        originalMemo: 'GITHUB.COM 877-448-4820 CA',
        members: [],
        createdAt: '2026-01-15T06:00:00Z',
      });
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01hy4rkq0x3c6dtr9w1p2v5bns')
      .reply(200, {
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        fundingScheduleId: 'fund_01hym37k3kj4ghv67nfx7vkvr0',
        spendingType: 'expense',
        name: 'Github Copilot',
        targetAmount: 800,
        currentAmount: 300,
        usedAmount: 0,
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
        lastRecurrence: null,
        nextRecurrence: '2099-04-15T05:00:00Z',
        nextContributionAmount: 250,
        isBehind: false,
        isPaused: false,
        autoCreateTransaction: false,
        createdAt: '2026-03-16T06:00:00Z',
      });
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
      fundingScheduleId: null,
      fundingSchedule: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2099-04-15T05:00:00Z',
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
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSummaryCard onMarkNotRecurring={() => {}} recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    // Only $3 is set aside, but it isn't behind so it will have the rest by the time the charge comes in
    await waitFor(() =>
      expect(world.getByTestId('recurring-coverage')).toHaveTextContent('Github Copilot will have $8.00 ready'),
    );
    await waitFor(() => expect(world.getByTestId('recurring-coverage')).toHaveAttribute('data-covered', 'true'));
  });

  it('will say the next charge is not covered when the expense is behind', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/similar/tcl_01hy4rf0p7mz9w2q3c4v5b6n7m')
      .reply(200, {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Github',
        originalMemo: 'GITHUB.COM 877-448-4820 CA',
        members: [],
        createdAt: '2026-01-15T06:00:00Z',
      });
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01hy4rkq0x3c6dtr9w1p2v5bns')
      .reply(200, {
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        fundingScheduleId: 'fund_01hym37k3kj4ghv67nfx7vkvr0',
        spendingType: 'expense',
        name: 'Github Copilot',
        targetAmount: 800,
        currentAmount: 300,
        usedAmount: 0,
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
        lastRecurrence: null,
        nextRecurrence: '2099-04-15T05:00:00Z',
        nextContributionAmount: 250,
        isBehind: true,
        isPaused: false,
        autoCreateTransaction: false,
        createdAt: '2026-03-16T06:00:00Z',
      });
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
      fundingScheduleId: null,
      fundingSchedule: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-15T05:00:00Z',
      next: '2099-04-15T05:00:00Z',
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
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSummaryCard onMarkNotRecurring={() => {}} recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() =>
      expect(world.getByTestId('recurring-coverage')).toHaveTextContent('Github Copilot only has $3.00 of $8.00'),
    );
    await waitFor(() => expect(world.getByTestId('recurring-coverage')).toHaveAttribute('data-covered', 'false'));
  });

  it('will not show the next charge once it has ended', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/similar/tcl_01hy4rf0p7mz9w2q3c4v5b6n7m')
      .reply(200, {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Netflix',
        originalMemo: 'NETFLIX.COM 866-579-7172 CA',
        members: [],
        createdAt: '2025-01-15T06:00:00Z',
      });
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      spendingId: null,
      fundingScheduleId: null,
      fundingSchedule: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20250101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      first: '2025-01-15T06:00:00Z',
      last: '2025-06-15T05:00:00Z',
      next: '2025-07-15T05:00:00Z',
      ended: true,
      confidence: 0.9,
      direction: 'debit',
      amounts: {
        1549: 6,
      },
      lastAmount: 1549,
      autoMatched: false,
      autoAssign: false,
      createdAt: '2025-03-15T06:00:00Z',
      updatedAt: '2025-08-15T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSummaryCard onMarkNotRecurring={() => {}} recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('recurring-status')).toHaveTextContent('Ended'));
    expect(world.queryByTestId('recurring-next')).not.toBeInTheDocument();
  });
});
