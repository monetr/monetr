import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import TransactionRecurringCard from '@monetr/interface/components/transactions/TransactionRecurringCard';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import Transaction from '@monetr/interface/models/Transaction';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('transaction recurring card', () => {
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

  it('will look up the expense it is budgeted with', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
        fundingScheduleId: null,
        window: 'monthly',
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

    const transaction = new Transaction({
      transactionId: ID.from<Transaction>('txn_01hy4rh2k8m3n4p5q6r7s8t9v0'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      amount: 800,
      spendingId: null,
      spendingAmount: null,
      createdBySpendingId: null,
      createdByFundingScheduleId: null,
      categories: [],
      date: '2026-03-15T05:00:00Z',
      authorizedDate: null,
      name: 'Github',
      originalName: 'GITHUB.COM 877-448-4820 CA',
      merchantName: null,
      originalMerchantName: null,
      isPending: false,
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      createdAt: '2026-03-15T06:00:00Z',
    });

    const world = testRenderer(<TransactionRecurringCard transaction={transaction} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions/txn_01hy4rh2k8m3n4p5q6r7s8t9v0/details',
    });

    // The recurring transaction only has the id, so the name has to come from looking the expense up
    await waitFor(() =>
      expect(world.getByTestId('transaction-recurring-card')).toHaveTextContent('Budgeted with Github Copilot'),
    );
  });

  it('will turn on auto assign', async () => {
    mockFetch
      .onPatch('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
        fundingScheduleId: null,
        window: 'monthly',
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
        autoAssign: true,
        createdAt: '2026-03-15T06:00:00Z',
        updatedAt: '2026-03-15T06:00:00Z',
      });
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
        fundingScheduleId: null,
        window: 'monthly',
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

    const transaction = new Transaction({
      transactionId: ID.from<Transaction>('txn_01hy4rh2k8m3n4p5q6r7s8t9v0'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      amount: 800,
      spendingId: null,
      spendingAmount: null,
      createdBySpendingId: null,
      createdByFundingScheduleId: null,
      categories: [],
      date: '2026-03-15T05:00:00Z',
      authorizedDate: null,
      name: 'Github',
      originalName: 'GITHUB.COM 877-448-4820 CA',
      merchantName: null,
      originalMerchantName: null,
      isPending: false,
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      createdAt: '2026-03-15T06:00:00Z',
    });

    const user = userEvent.setup();
    const world = testRenderer(<TransactionRecurringCard transaction={transaction} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions/txn_01hy4rh2k8m3n4p5q6r7s8t9v0/details',
    });

    await waitFor(() => expect(world.getByTestId('transaction-recurring-auto-assign')).toBeEnabled());
    await user.click(world.getByTestId('transaction-recurring-auto-assign'));

    await waitFor(() => expect(mockFetch.history.patch).toHaveLength(1));
    expect(mockFetch.history.patch?.[0]?.data).toEqual({
      autoAssign: true,
    });
  });

  it('will say where the last one was spent from when nothing is linked', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: null,
        fundingScheduleId: null,
        window: 'monthly',
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
      });
    // Only the most recent one gets asked for
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=1',
      )
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v0',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
          spendingId: null,
          amount: 800,
          date: '2026-03-15T05:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM 877-448-4820 CA',
          isPending: false,
          createdAt: '2026-03-15T06:00:00Z',
        },
      ]);
    apiSampleResponses(mockFetch);

    const transaction = new Transaction({
      transactionId: ID.from<Transaction>('txn_01hy4rh2k8m3n4p5q6r7s8t9v0'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      amount: 800,
      spendingId: null,
      spendingAmount: null,
      createdBySpendingId: null,
      createdByFundingScheduleId: null,
      categories: [],
      date: '2026-03-15T05:00:00Z',
      authorizedDate: null,
      name: 'Github',
      originalName: 'GITHUB.COM 877-448-4820 CA',
      merchantName: null,
      originalMerchantName: null,
      isPending: false,
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      createdAt: '2026-03-15T06:00:00Z',
    });

    const world = testRenderer(<TransactionRecurringCard transaction={transaction} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions/txn_01hy4rh2k8m3n4p5q6r7s8t9v0/details',
    });

    await waitFor(() =>
      expect(world.getByTestId('transaction-recurring-card')).toHaveTextContent('Last spent from Free-To-Use'),
    );
  });
});
