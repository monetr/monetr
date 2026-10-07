import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import Recurring from '@monetr/interface/pages/recurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('recurring page', () => {
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

  it('will show the empty state when nothing is recurring', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring?direction=debit&ended=false')
      .reply(200, []);
    apiSampleResponses(mockFetch);

    const world = testRenderer(<Recurring />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring',
    });

    await waitFor(() => expect(world.getByTestId('recurring-empty')).toHaveTextContent('Nothing here yet...'));
  });

  it('will split things up into tabs', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring?direction=debit&ended=false')
      .reply(200, [
        {
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx30201',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n01',
          transactionCluster: {
            transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n01',
            bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
            name: 'Github',
            originalMemo: 'GITHUB.COM 877-448-4820 CA',
            createdAt: '2026-01-15T06:00:00Z',
          },
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
          createdAt: '2026-03-15T06:00:00Z',
          updatedAt: '2026-03-15T06:00:00Z',
        },
      ]);
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring?direction=credit&ended=false')
      .reply(200, [
        {
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx30202',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n02',
          transactionCluster: {
            transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n02',
            bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
            name: 'Mercury Payroll',
            originalMemo: 'MERCURY PAYROLL DIRECT DEP',
            createdAt: '2026-01-15T06:00:00Z',
          },
          spendingId: null,
          fundingScheduleId: null,
          window: 'monthly',
          ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
          first: '2026-01-15T06:00:00Z',
          last: '2026-03-15T05:00:00Z',
          next: '2099-04-15T05:00:00Z',
          ended: false,
          confidence: 0.9,
          direction: 'credit',
          amounts: {
            '-250000': 3,
          },
          lastAmount: -250000,
          autoMatched: false,
          createdAt: '2026-03-15T06:00:00Z',
          updatedAt: '2026-03-15T06:00:00Z',
        },
      ]);
    mockFetch.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring?ended=true').reply(200, [
      {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx30203',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n03',
        transactionCluster: {
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n03',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          name: 'Netflix',
          originalMemo: 'NETFLIX.COM 866-579-7172 CA',
          createdAt: '2026-01-15T06:00:00Z',
        },
        spendingId: null,
        fundingScheduleId: null,
        window: 'monthly',
        ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
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
        createdAt: '2025-03-15T06:00:00Z',
        updatedAt: '2025-08-15T06:00:00Z',
      },
    ]);
    apiSampleResponses(mockFetch);

    const user = userEvent.setup();
    const world = testRenderer(<Recurring />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring',
    });

    // Charges are the default tab
    await waitFor(() => expect(world.getAllByTestId('recurring-item-name')).toHaveLength(1));
    await waitFor(() => expect(world.getByTestId('recurring-item-name')).toHaveTextContent('Github'));

    await user.click(world.getByTestId('recurring-tab-deposits'));
    await waitFor(() => expect(world.getByTestId('recurring-item-name')).toHaveTextContent('Mercury Payroll'));

    await user.click(world.getByTestId('recurring-tab-ended'));
    await waitFor(() => expect(world.getByTestId('recurring-item-name')).toHaveTextContent('Netflix'));
  });

  it('will put things that were due a while ago in the late group', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring?direction=debit&ended=false')
      .reply(200, [
        {
          transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx30201',
          bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
          transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n01',
          transactionCluster: {
            transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n01',
            bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
            name: 'Hulu',
            originalMemo: 'HULU 877-8244858 CA',
            createdAt: '2020-01-15T06:00:00Z',
          },
          spendingId: null,
          fundingScheduleId: null,
          window: 'monthly',
          ruleset: 'DTSTART:20200101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
          first: '2020-01-15T06:00:00Z',
          last: '2020-03-15T05:00:00Z',
          next: '2020-04-15T05:00:00Z',
          ended: false,
          confidence: 0.9,
          direction: 'debit',
          amounts: {
            1799: 3,
          },
          lastAmount: 1799,
          autoMatched: false,
          createdAt: '2020-03-15T06:00:00Z',
          updatedAt: '2020-03-15T06:00:00Z',
        },
      ]);
    apiSampleResponses(mockFetch);

    const world = testRenderer(<Recurring />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring',
    });

    await waitFor(() => expect(world.getByTestId('recurring-group-late')).toHaveTextContent('Late'));
    await waitFor(() => expect(world.getByTestId('recurring-item-name')).toHaveTextContent('Hulu'));
  });
});
