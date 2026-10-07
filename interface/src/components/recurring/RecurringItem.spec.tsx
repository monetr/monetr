import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import RecurringItem from '@monetr/interface/components/recurring/RecurringItem';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('recurring item', () => {
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

  it('will show the name and a create button when nothing is linked', async () => {
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      transactionCluster: {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Github',
        originalMemo: 'GITHUB.COM 877-448-4820 CA',
        members: [],
        createdAt: '2026-01-15T06:00:00Z',
      },
      spendingId: null,
      spending: null,
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
    });

    const world = testRenderer(<RecurringItem recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring',
    });

    await waitFor(() => expect(world.getByTestId('recurring-item-name')).toHaveTextContent('Github'));
    await waitFor(() => expect(world.getByTestId('recurring-item-create')).toHaveTextContent('Create'));
    expect(world.queryByTestId('recurring-item-linked')).not.toBeInTheDocument();
  });

  it('will look up the linked expense', async () => {
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
      transactionCluster: {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Github',
        originalMemo: 'GITHUB.COM 877-448-4820 CA',
        members: [],
        createdAt: '2026-01-15T06:00:00Z',
      },
      spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
      spending: null,
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
    });

    const world = testRenderer(<RecurringItem recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring',
    });

    await waitFor(() => expect(world.getByTestId('recurring-item-linked')).toHaveTextContent('Github Copilot'));
    expect(world.queryByTestId('recurring-item-create')).not.toBeInTheDocument();
    // The expense isn't behind so it isn't short, even though it doesn't have enough in it yet
    expect(world.queryByTestId('recurring-item-short')).not.toBeInTheDocument();
  });

  it('will show how short a behind expense is', async () => {
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
      transactionCluster: {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Github',
        originalMemo: 'GITHUB.COM 877-448-4820 CA',
        members: [],
        createdAt: '2026-01-15T06:00:00Z',
      },
      spendingId: ID.from<Spending>('spnd_01hy4rkq0x3c6dtr9w1p2v5bns'),
      spending: null,
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
    });

    const world = testRenderer(<RecurringItem recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring',
    });

    // $8.00 charge with $3.00 set aside
    await waitFor(() => expect(world.getByTestId('recurring-item-short')).toHaveTextContent('short $5.00'));
  });

  it('will show the recent charges when expanded', async () => {
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=100',
      )
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v0',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          amount: 800,
          date: '2026-03-15T05:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM 877-448-4820 CA MARCH',
          isPending: false,
          createdAt: '2026-03-15T06:00:00Z',
        },
      ]);
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      transactionCluster: {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        name: 'Github',
        originalMemo: 'GITHUB.COM 877-448-4820 CA',
        members: [],
        createdAt: '2026-01-15T06:00:00Z',
      },
      spendingId: null,
      spending: null,
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
    });

    const user = userEvent.setup();
    const world = testRenderer(<RecurringItem recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring',
    });

    await waitFor(() => expect(world.getByTestId('recurring-item-toggle')).toBeInTheDocument());
    // Open up the recent charges
    await user.click(world.getByTestId('recurring-item-toggle'));

    await waitFor(() => expect(world.getByTitle('GITHUB.COM 877-448-4820 CA MARCH')).toBeInTheDocument());
    await waitFor(() => expect(world.getByTestId('recurring-item-toggle')).toHaveAttribute('aria-expanded', 'true'));
  });
});
