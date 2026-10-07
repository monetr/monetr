import { waitFor } from '@testing-library/react';

import RecurringClusterCard from '@monetr/interface/components/recurring/RecurringClusterCard';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('recurring cluster card', () => {
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

  it('will count up the different ways the charges show up', async () => {
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_cluster_id=tcl_01hy4rf0p7mz9w2q3c4v5b6n7m&limit=100',
      )
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v1',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
          amount: 299,
          date: '2026-03-11T05:00:00Z',
          name: 'Apple',
          originalName: 'APPLE.COM/BILL 866-712-7753 CA',
          isPending: false,
          createdAt: '2026-03-11T06:00:00Z',
        },
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v2',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
          amount: 999,
          date: '2026-03-02T05:00:00Z',
          name: 'Apple',
          originalName: 'APPLE.COM/BILL ITUNES.COM',
          isPending: false,
          createdAt: '2026-03-02T06:00:00Z',
        },
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v3',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
          amount: 299,
          date: '2026-02-11T05:00:00Z',
          name: 'Apple',
          originalName: 'APPLE.COM/BILL 866-712-7753 CA',
          isPending: false,
          createdAt: '2026-02-11T06:00:00Z',
        },
      ]);
    mockFetch
      .onGet(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions?transaction_recurring_id=txrc_01hy4re7c1xc2v44cf6kx302jx&limit=100',
      )
      .reply(200, []);
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      transactionCluster: null,
      spendingId: null,
      spending: null,
      fundingScheduleId: null,
      fundingSchedule: null,
      window: TransactionRecurringWindow.Monthly,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=11',
      first: '2026-01-11T06:00:00Z',
      last: '2026-03-11T05:00:00Z',
      next: '2099-04-11T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'debit',
      amounts: {
        299: 9,
      },
      lastAmount: 299,
      autoMatched: false,
      createdAt: '2026-03-11T06:00:00Z',
      updatedAt: '2026-03-11T06:00:00Z',
    });

    const world = testRenderer(<RecurringClusterCard name='Apple' recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    // Most common one first
    await waitFor(() =>
      expect(world.getByTestId('recurring-memos')).toHaveTextContent(
        'APPLE.COM/BILL 866-712-7753 CA2 chargesAPPLE.COM/BILL ITUNES.COM1 charge',
      ),
    );
    // How many are on the schedule comes from the amounts on the recurring transaction
    await waitFor(() => expect(world.getByTestId('recurring-on-schedule')).toHaveTextContent('9 on this schedule'));
  });
});
