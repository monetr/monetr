import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import RecurringChargeList from '@monetr/interface/components/recurring/RecurringChargeList';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('recurring charge list', () => {
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

  it('will only load the charges on the schedule to start', async () => {
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=10',
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
    });

    const world = testRenderer(<RecurringChargeList recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTitle('GITHUB.COM 877-448-4820 CA MARCH')).toBeInTheDocument());
    // Plenty of people never look at the all tab, so the whole group shouldn't be requested until they do
    expect(mockFetch.history.get?.map(item => item.url)).not.toContain(
      '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_cluster_id=tcl_01hy4rf0p7mz9w2q3c4v5b6n7m&limit=10',
    );
  });

  it('will load the whole group when switching to all', async () => {
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=10',
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
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_cluster_id=tcl_01hy4rf0p7mz9w2q3c4v5b6n7m&limit=10',
      )
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v1',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionRecurringId: null,
          amount: 2000,
          date: '2026-03-20T05:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM SPONSORS 877-448-4820 CA',
          isPending: false,
          createdAt: '2026-03-20T06:00:00Z',
        },
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
    });

    const user = userEvent.setup();
    const world = testRenderer(<RecurringChargeList recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTitle('GITHUB.COM 877-448-4820 CA MARCH')).toBeInTheDocument());
    await user.click(world.getByTestId('recurring-charges-tab-all'));

    // The sponsors charge is in the group but not on the schedule
    await waitFor(() => expect(world.getByTitle('GITHUB.COM SPONSORS 877-448-4820 CA')).toBeInTheDocument());
    await waitFor(() => expect(world.getAllByTestId('recurring-charge-one-off')).toHaveLength(1));
  });
});
