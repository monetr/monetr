import { waitFor } from '@testing-library/react';

import RecurringFundingCard from '@monetr/interface/components/recurring/RecurringFundingCard';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('recurring funding card', () => {
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

  it('will disable the select when there are no funding schedules', async () => {
    // This has to be registered before the sample responses so it wins over their funding schedules
    mockFetch.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules').reply(200, []);
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      spendingId: null,
      fundingScheduleId: null,
      fundingSchedule: null,
      window: TransactionRecurringWindow.FifteenthAndLast,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-31T05:00:00Z',
      next: '2099-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'credit',
      amounts: {
        '-250000': 6,
      },
      lastAmount: -250000,
      autoMatched: false,
      createdAt: '2026-03-31T06:00:00Z',
      updatedAt: '2026-03-31T06:00:00Z',
    });

    const world = testRenderer(<RecurringFundingCard name='Mercury Payroll' recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    // The select doesn't take a test id, so find it by its placeholder
    await waitFor(() => expect(world.getByPlaceholderText('No funding schedules exist...')).toBeDisabled());
    await waitFor(() => expect(world.getByTestId('recurring-new-funding')).toBeVisible());
  });

  it('will link to the funding schedule when one is linked', async () => {
    apiSampleResponses(mockFetch);

    const recurring = new TransactionRecurring({
      transactionRecurringId: ID.from<TransactionRecurring>('txrc_01hy4re7c1xc2v44cf6kx302jx'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      transactionClusterId: ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m'),
      spendingId: null,
      fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
      fundingSchedule: {
        fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
        bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
        name: "Elliot's Contribution",
        description: '15th and last day of every month',
        ruleset: 'DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        excludeWeekends: true,
        autoCreateTransaction: false,
        estimatedDeposit: 22000,
        lastRecurrence: '2024-08-15T05:00:00Z',
        nextRecurrence: '2024-08-30T05:00:00Z',
        nextRecurrenceOriginal: '2024-08-31T05:00:00Z',
      },
      window: TransactionRecurringWindow.FifteenthAndLast,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
      first: '2026-01-15T06:00:00Z',
      last: '2026-03-31T05:00:00Z',
      next: '2099-04-15T05:00:00Z',
      ended: false,
      confidence: 0.9,
      direction: 'credit',
      amounts: {
        '-250000': 6,
      },
      lastAmount: -250000,
      autoMatched: false,
      createdAt: '2026-03-31T06:00:00Z',
      updatedAt: '2026-03-31T06:00:00Z',
    });

    const world = testRenderer(<RecurringFundingCard name='Mercury Payroll' recurring={recurring} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx/details',
    });

    await waitFor(() => expect(world.getByTestId('recurring-view-funding')).toBeVisible());
    expect(world.queryByTestId('recurring-new-funding')).not.toBeInTheDocument();
  });
});
