import { useFormikContext } from 'formik';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import MForm from '@monetr/interface/components/MForm';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import type BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import Transaction from '@monetr/interface/models/Transaction';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('select spending', () => {
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

  it('will offer to create an expense from a recurring charge', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: null,
        fundingScheduleId: null,
        fundingSchedule: null,
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
        deletedAt: null,
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
    const world = testRenderer(
      <MForm
        initialValues={{
          spendingId: null,
        }}
        onSubmit={() => {}}
      >
        <MSelectSpending name='spendingId' transaction={transaction} />
      </MForm>,
      { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions/txn_01hy4rh2k8m3n4p5q6r7s8t9v0/details' },
    );

    // Open the select once everything has loaded
    await waitFor(() => expect(world.getByRole('combobox')).toBeEnabled());
    await user.click(world.getByRole('combobox'));
    await waitFor(() => expect(world.getAllByRole('option')[0]).toHaveTextContent('Create Expense From Github'));
    await waitFor(() => expect(world.getByText('This charge repeats every month on the 15th')).toBeVisible());
    expect(world.queryByText('Suggested')).not.toBeInTheDocument();
  });

  it('will suggest the expense the recurring charges are linked to', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01h264znvxkghmxp5s0wmdbr9j',
        fundingScheduleId: null,
        fundingSchedule: null,
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
        deletedAt: null,
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
    const world = testRenderer(
      <MForm
        initialValues={{
          spendingId: null,
        }}
        onSubmit={() => {}}
      >
        <MSelectSpending name='spendingId' transaction={transaction} />
      </MForm>,
      { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions/txn_01hy4rh2k8m3n4p5q6r7s8t9v0/details' },
    );

    // Open the select once everything has loaded
    await waitFor(() => expect(world.getByRole('combobox')).toBeEnabled());
    await user.click(world.getByRole('combobox'));
    await waitFor(() => expect(world.getAllByRole('option')[0]).toHaveTextContent('GitLab'));
    expect(world.getAllByRole('option')[0]).toHaveTextContent('Linked to your Github charges');
    expect(world.getByText('Suggested')).toBeVisible();
    // GitLab moved to the top, so it should only be in the list once
    expect(world.getAllByRole('option').filter(option => option.textContent?.includes('GitLab'))).toHaveLength(1);
    expect(world.queryByText('Create Expense From Github')).not.toBeInTheDocument();
  });

  it('will not suggest the expense it is already spent from', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01h264znvxkghmxp5s0wmdbr9j',
        fundingScheduleId: null,
        fundingSchedule: null,
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
        deletedAt: null,
      });
    apiSampleResponses(mockFetch);

    const transaction = new Transaction({
      transactionId: ID.from<Transaction>('txn_01hy4rh2k8m3n4p5q6r7s8t9v0'),
      bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
      amount: 800,
      spendingId: ID.from<Spending>('spnd_01h264znvxkghmxp5s0wmdbr9j'),
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
    const world = testRenderer(
      <MForm
        initialValues={{
          spendingId: 'spnd_01h264znvxkghmxp5s0wmdbr9j',
        }}
        onSubmit={() => {}}
      >
        <MSelectSpending name='spendingId' transaction={transaction} />
      </MForm>,
      { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions/txn_01hy4rh2k8m3n4p5q6r7s8t9v0/details' },
    );

    // Wait for the recurring transaction too, otherwise there would never be a suggestion to begin with
    await waitFor(() => expect(world.getByRole('combobox')).toHaveValue('GitLab'));
    await waitFor(() =>
      expect(mockFetch.history.get?.map(item => item.url)).toContain(
        '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx',
      ),
    );
    await user.click(world.getByRole('combobox'));
    await waitFor(() => expect(world.getAllByRole('option')[0]).toHaveTextContent('Free-To-Use'));
    expect(world.queryByText('Suggested')).not.toBeInTheDocument();
  });

  it('will pick the suggested expense', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/recurring/txrc_01hy4re7c1xc2v44cf6kx302jx')
      .reply(200, {
        transactionRecurringId: 'txrc_01hy4re7c1xc2v44cf6kx302jx',
        bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        spendingId: 'spnd_01h264znvxkghmxp5s0wmdbr9j',
        fundingScheduleId: null,
        fundingSchedule: null,
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
        autoMatched: true,
        autoAssign: false,
        createdAt: '2026-03-15T06:00:00Z',
        updatedAt: '2026-03-15T06:00:00Z',
        deletedAt: null,
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
    const world = testRenderer(
      <MForm
        initialValues={{
          spendingId: null,
        }}
        onSubmit={() => {}}
      >
        <MSelectSpending name='spendingId' transaction={transaction} />
        <SpendingValue />
      </MForm>,
      { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions/txn_01hy4rh2k8m3n4p5q6r7s8t9v0/details' },
    );

    // Open the select once everything has loaded
    await waitFor(() => expect(world.getByRole('combobox')).toBeEnabled());
    await user.click(world.getByRole('combobox'));
    await waitFor(() =>
      expect(world.getAllByRole('option')[0]).toHaveTextContent(
        'monetr picked this since your recent Github charges were spent from it',
      ),
    );
    await user.click(world.getByText('GitLab'));
    await waitFor(() =>
      expect(world.getByTestId('spending-value')).toHaveTextContent('spnd_01h264znvxkghmxp5s0wmdbr9j'),
    );
  });
});

function SpendingValue(): React.JSX.Element {
  const { values } = useFormikContext<{ spendingId: string | null }>();
  return <span data-testid='spending-value'>{values.spendingId}</span>;
}
