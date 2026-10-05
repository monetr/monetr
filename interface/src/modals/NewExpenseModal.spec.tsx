import { act } from 'react';

import { waitFor } from '@testing-library/react';

import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import Transaction from '@monetr/interface/models/Transaction';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import TransactionRecurring, { TransactionRecurringWindow } from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

const bankAccountId = 'bac_01gds6eqsq7h5mgevwtmw3cyxb';
const clusterId = 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m';
const recurringId = 'txrc_01hy4re7c1xc2v44cf6kx302jx';
const transactionId = 'txn_01hy4rhqmy4wjy0vtrmqsc5c1m';
const fundingScheduleId = 'fund_01hy4re7c1xc2v44cf6kx302jx';

function recurring(next: string): TransactionRecurring {
  return new TransactionRecurring({
    transactionRecurringId: ID.from<TransactionRecurring>(recurringId),
    bankAccountId: ID.from<BankAccount>(bankAccountId),
    transactionClusterId: ID.from<TransactionCluster>(clusterId),
    window: TransactionRecurringWindow.Monthly,
    ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
    first: '2026-01-15T06:00:00Z',
    last: '2026-03-15T05:00:00Z',
    next,
    ended: false,
    confidence: 0.9,
    direction: 'debit',
    amounts: { 1549: 3 },
    lastAmount: 1549,
    spending: null,
    fundingSchedule: null,
    createdAt: '2026-01-15T06:00:00Z',
    updatedAt: '2026-03-16T06:00:00Z',
  });
}

const transaction = new Transaction({
  transactionId: ID.from<Transaction>(transactionId),
  bankAccountId: ID.from<BankAccount>(bankAccountId),
  amount: 1549,
  spendingId: null,
  spendingAmount: null,
  createdBySpendingId: null,
  createdByFundingScheduleId: null,
  categories: [],
  date: '2026-03-15T05:00:00Z',
  authorizedDate: null,
  // different from the cluster name on purpose, the expense should use the cluster name
  name: 'NETFLIX.COM 866-579-7172 CA',
  originalName: 'NETFLIX.COM 866-579-7172 CA',
  merchantName: null,
  originalMerchantName: null,
  isPending: false,
  transactionClusterId: ID.from<TransactionCluster>(clusterId),
  transactionRecurringId: ID.from<TransactionRecurring>(recurringId),
  createdAt: '2026-03-15T05:00:00Z',
});

function givenTheRecurringMocks(mockFetch: FetchMock) {
  mockFetch.onGet(`/api/bank_accounts/${bankAccountId}`).reply(200, {
    bankAccountId,
    linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
    availableBalance: 47986,
    currentBalance: 47986,
    mask: '2982',
    name: 'Mercury Checking',
    originalName: 'Mercury Checking',
    accountType: 'depository',
    accountSubType: 'checking',
    status: 'active',
    lastUpdated: '2024-08-27T08:53:48.555368Z',
    createdAt: '2022-09-25T02:08:40.758642Z',
    updatedAt: '2024-03-19T06:17:32.335106Z',
  });
  mockFetch.onGet(`/api/bank_accounts/${bankAccountId}/funding_schedules`).reply(200, [
    {
      bankAccountId,
      dateStarted: '2023-02-28T06:00:00Z',
      description: '15th and last day of every month',
      estimatedDeposit: null,
      excludeWeekends: false,
      fundingScheduleId,
      lastRecurrence: '2023-09-29T05:00:00Z',
      name: 'Payday',
      nextRecurrence: '2023-10-13T05:00:00Z',
      nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
      ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
      waitForDeposit: false,
    },
  ]);
  mockFetch.onGet(`/api/bank_accounts/${bankAccountId}/similar/${clusterId}`).reply(200, {
    transactionClusterId: clusterId,
    bankAccountId,
    name: 'Netflix',
    originalName: 'NETFLIX.COM',
    createdAt: '2026-01-15T06:00:00Z',
    updatedAt: '2026-03-15T06:00:00Z',
  });
  mockFetch
    .onGet(`/api/bank_accounts/${bankAccountId}/transactions?transaction_recurring_id=${recurringId}&limit=100`)
    .reply(200, []);
}

describe('new expense modal', () => {
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

  it('will render', async () => {
    mockFetch.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
      bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
      availableBalance: 47986,
      currentBalance: 47986,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      lastUpdated: '2024-08-27T08:53:48.555368Z',
      createdAt: '2022-09-25T02:08:40.758642Z',
      updatedAt: '2024-03-19T06:17:32.335106Z',
    });
    mockFetch.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules').reply(200, []);

    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses' });
    // Open the dialog
    await act(() => void showNewExpenseModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-expense-modal')).toBeVisible());
    // Close the dialog.
    act(() => world.getByTestId('close-new-expense-modal').click());
    // Make sure it goes away.
    await waitFor(() => expect(world.queryByTestId('new-expense-modal')).not.toBeInTheDocument());
  });

  it('will fill in from a recurring transaction', async () => {
    givenTheRecurringMocks(mockFetch);

    const world = testRenderer(<div />, { initialRoute: `/bank/${bankAccountId}/transactions` });
    await act(() => void showNewExpenseModal({ recurring: recurring('2099-04-15T05:00:00Z'), transaction }));
    await waitFor(() => expect(world.getByTestId('new-expense-modal')).toBeVisible());

    expect(world.getByTestId('new-expense-recurring-banner')).toHaveTextContent(
      'Filled in from your 3 Netflix charges',
    );
    // the cluster name wins over the transactions own name
    expect(world.getByDisplayValue('Netflix')).toBeInTheDocument();
    expect(world.getByDisplayValue('$15.49')).toBeInTheDocument();
    await waitFor(() => expect(world.getByDisplayValue('Every month on the 15th')).toBeInTheDocument());
    // theres only one funding schedule so it should already be picked
    await waitFor(() => expect(world.getByDisplayValue('Payday')).toBeInTheDocument());
  });

  it('will create the expense and spend the charge from it', async () => {
    givenTheRecurringMocks(mockFetch);
    mockFetch.onPost(`/api/bank_accounts/${bankAccountId}/spending`).reply(200, {
      spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
      bankAccountId,
      fundingScheduleId,
      transactionRecurringId: recurringId,
      name: 'Netflix',
      spendingType: 'expense',
      targetAmount: 1549,
      currentAmount: 0,
      usedAmount: 0,
      ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
      lastRecurrence: null,
      nextRecurrence: '2099-04-15T05:00:00Z',
      nextContributionAmount: 774,
      isBehind: false,
      isPaused: false,
      autoCreateTransaction: false,
      createdAt: '2026-03-16T06:00:00Z',
    });
    mockFetch.onPatch(`/api/bank_accounts/${bankAccountId}/transactions/${transactionId}`).reply(200, {
      transaction: {
        ...transaction,
        spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns',
      },
      spending: [],
      balance: {
        bankAccountId,
        available: 47986,
        current: 47986,
        limit: 0,
        free: 46437,
        expenses: 0,
        goals: 0,
      },
    });

    const world = testRenderer(<div />, { initialRoute: `/bank/${bankAccountId}/transactions` });
    // next is in the past here, so the expense should be due on the next 15th instead
    await act(() => void showNewExpenseModal({ recurring: recurring('2020-01-15T06:00:00Z'), transaction }));
    await waitFor(() => expect(world.getByTestId('new-expense-modal')).toBeVisible());
    await waitFor(() => expect(world.getByDisplayValue('Payday')).toBeInTheDocument());
    await waitFor(() => expect(world.getByDisplayValue('Every month on the 15th')).toBeInTheDocument());

    // spending the charge from the new expense is on by default
    expect(world.getByTestId('new-expense-move-transaction')).toBeChecked();
    act(() => world.getByRole('button', { name: 'Create' }).click());

    await waitFor(() => expect(mockFetch.history.patch).toHaveLength(1));
    expect(mockFetch.history.post).toHaveLength(1);
    const created = mockFetch.history.post?.[0]?.data as Record<string, unknown>;
    expect(created.transactionRecurringId).toBe(recurringId);
    expect(created.name).toBe('Netflix');
    expect(created.targetAmount).toBe(1549);
    expect(created.fundingScheduleId).toBe(fundingScheduleId);
    expect(new Date(created.nextRecurrence as string).getTime()).toBeGreaterThan(Date.now());
    const patched = mockFetch.history.patch?.[0];
    expect(patched?.url).toBe(`/api/bank_accounts/${bankAccountId}/transactions/${transactionId}`);
    expect(patched?.data).toMatchObject({ spendingId: 'spnd_01hy4rkq0x3c6dtr9w1p2v5bns' });
    await waitFor(() => expect(world.queryByTestId('new-expense-modal')).not.toBeInTheDocument());
  });
});
