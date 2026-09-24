import { waitFor } from '@testing-library/react';

import ExpenseItem from '@monetr/interface/components/expenses/ExpenseItem';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('expense item', () => {
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

  it('will show the contribution for an expense', async () => {
    apiSampleResponses(mockFetch);

    const spending = new Spending({
      spendingId: ID.from<Spending>('spnd_01fpwx2qng9gbt32kdh2mdk7df'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
      spendingType: SpendingType.Expense,
      name: 'Freshbooks',
      targetAmount: 1900,
      currentAmount: 950,
      usedAmount: 0,
      ruleset: 'DTSTART:20230310T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10',
      lastRecurrence: '2024-08-10T06:00:00Z',
      nextRecurrence: '2024-09-10T06:00:00Z',
      nextContributionAmount: 950,
      isBehind: false,
      autoCreateTransaction: false,
      isPaused: false,
      createdAt: '2021-12-14T16:40:46Z',
    });

    const world = testRenderer(<ExpenseItem spending={spending} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses',
    });

    await waitFor(() => expect(world.getAllByText("$9.50 / Elliot's Contribution")).not.toHaveLength(0));
    expect(world.queryByText('Paused')).not.toBeInTheDocument();
  });

  it('will show paused instead of the contribution for a paused expense', async () => {
    apiSampleResponses(mockFetch);

    const spending = new Spending({
      spendingId: ID.from<Spending>('spnd_01fpwx2qng9gbt32kdh2mdk7df'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
      spendingType: SpendingType.Expense,
      name: 'Freshbooks',
      targetAmount: 1900,
      currentAmount: 950,
      usedAmount: 0,
      ruleset: 'DTSTART:20230310T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10',
      lastRecurrence: '2024-08-10T06:00:00Z',
      nextRecurrence: '2024-09-10T06:00:00Z',
      nextContributionAmount: 950,
      isBehind: false,
      autoCreateTransaction: false,
      isPaused: true,
      createdAt: '2021-12-14T16:40:46Z',
    });

    const world = testRenderer(<ExpenseItem spending={spending} />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses',
    });

    await waitFor(() => expect(world.getAllByText('Paused')).not.toHaveLength(0));
    expect(world.queryByText("$9.50 / Elliot's Contribution")).not.toBeInTheDocument();
  });
});
