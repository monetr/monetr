import { waitFor } from '@testing-library/react';

import RecurringSpendingCard from '@monetr/interface/components/recurring/RecurringSpendingCard';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('recurring spending card', () => {
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

  it('will disable the select when there are no expenses', async () => {
    // This has to be registered before the sample responses so it wins over their spending list
    mockFetch.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending').reply(200, []);
    apiSampleResponses(mockFetch);

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
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-15T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSpendingCard name='Github' recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    // The select doesn't take a test id, so find it by its placeholder
    await waitFor(() => expect(world.getByPlaceholderText('No expenses exist...')).toBeDisabled());
    await waitFor(() => expect(world.getByTestId('recurring-new-expense')).toBeVisible());
  });

  it('will not offer to create an expense when one is already linked', async () => {
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
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-15T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSpendingCard name='Github' recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('recurring-view-expense')).toBeVisible());
    expect(world.queryByTestId('recurring-new-expense')).not.toBeInTheDocument();
    expect(world.queryByTestId('recurring-auto-matched')).not.toBeInTheDocument();
  });

  it('will say when monetr linked it on its own', async () => {
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
      autoMatched: true,
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-15T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSpendingCard name='Github' recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('recurring-auto-matched')).toHaveTextContent('Github Copilot'));
  });

  it('will have the auto spend switch turned off for now', async () => {
    apiSampleResponses(mockFetch);

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
      createdAt: '2026-03-15T06:00:00Z',
      updatedAt: '2026-03-15T06:00:00Z',
      deletedAt: null,
    });

    const world = testRenderer(<RecurringSpendingCard name='Github' recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('recurring-auto-spend')).toBeDisabled());
  });
});
