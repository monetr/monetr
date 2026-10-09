import { act } from 'react';
import { rs } from '@rstest/core';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { showRemoveExpenseModal } from '@monetr/interface/modals/RemoveExpenseModal';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import Expenses from '@monetr/interface/pages/expenses';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('remove expense modal', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
    mockEnqueueSnackbar.mockReset();
  });
  afterEach(() => {
    mockFetch.reset();
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will remove the expense', async () => {
    apiSampleResponses(mockFetch);
    mockFetch
      .onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01h264znvxkghmxp5s0wmdbr9j')
      .reply(200);

    // Removing the expense prunes it from the cached spending list, so the list needs to be loaded first.
    const world = testRenderer(<Expenses />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveExpenseModal({
          spending: new Spending({
            spendingId: ID.from<Spending>('spnd_01h264znvxkghmxp5s0wmdbr9j'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            name: 'GitLab',
            spendingType: SpendingType.Expense,
            targetAmount: 1900,
            currentAmount: 1900,
            usedAmount: 0,
            ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
            lastRecurrence: null,
            nextRecurrence: '2099-04-15T05:00:00Z',
            nextContributionAmount: 0,
            isBehind: false,
            isPaused: false,
            autoCreateTransaction: false,
            createdAt: '2026-03-16T06:00:00Z',
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-expense-modal')).toBeVisible());
    // Nothing should be sent until the user confirms
    expect(mockFetch.history.delete).toHaveLength(0);

    await act(() => user.click(world.getByTestId('remove-expense-confirm')));

    await waitFor(() => expect(world.queryByTestId('remove-expense-modal')).not.toBeInTheDocument());
    await waitFor(() => expect(onRemoved).toHaveBeenCalled());
    expect(mockFetch.history.delete).toHaveLength(1);
    expect(mockFetch.history.delete?.[0]).toMatchObject({
      url: '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01h264znvxkghmxp5s0wmdbr9j',
    });
    expect(mockEnqueueSnackbar).not.toHaveBeenCalled();
  });

  it('will not remove the expense when cancelled', async () => {
    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveExpenseModal({
          spending: new Spending({
            spendingId: ID.from<Spending>('spnd_01h264znvxkghmxp5s0wmdbr9j'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            name: 'GitLab',
            spendingType: SpendingType.Expense,
            targetAmount: 1900,
            currentAmount: 1900,
            usedAmount: 0,
            ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
            lastRecurrence: null,
            nextRecurrence: '2099-04-15T05:00:00Z',
            nextContributionAmount: 0,
            isBehind: false,
            isPaused: false,
            autoCreateTransaction: false,
            createdAt: '2026-03-16T06:00:00Z',
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-expense-modal')).toBeVisible());

    await user.click(world.getByTestId('close-remove-expense-modal'));

    // Make sure it goes away without sending anything.
    await waitFor(() => expect(world.queryByTestId('remove-expense-modal')).not.toBeInTheDocument());
    expect(mockFetch.history.delete).toHaveLength(0);
    expect(onRemoved).not.toHaveBeenCalled();
  });

  it('will show an error when the expense cannot be removed', async () => {
    mockFetch
      .onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01h264znvxkghmxp5s0wmdbr9j')
      .reply(400, {
        error: 'Failed to remove spending',
      });

    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveExpenseModal({
          spending: new Spending({
            spendingId: ID.from<Spending>('spnd_01h264znvxkghmxp5s0wmdbr9j'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            name: 'GitLab',
            spendingType: SpendingType.Expense,
            targetAmount: 1900,
            currentAmount: 1900,
            usedAmount: 0,
            ruleset: 'DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15',
            lastRecurrence: null,
            nextRecurrence: '2099-04-15T05:00:00Z',
            nextContributionAmount: 0,
            isBehind: false,
            isPaused: false,
            autoCreateTransaction: false,
            createdAt: '2026-03-16T06:00:00Z',
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-expense-modal')).toBeVisible());

    await act(() => user.click(world.getByTestId('remove-expense-confirm')));

    // The modal stays open so the user can try again
    await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Failed to remove spending', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
    expect(world.getByTestId('remove-expense-modal')).toBeVisible();
    await waitFor(() => expect(world.getByTestId('remove-expense-confirm')).toBeEnabled());
    expect(onRemoved).not.toHaveBeenCalled();
  });
});
