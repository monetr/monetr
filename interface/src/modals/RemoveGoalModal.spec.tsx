import { act } from 'react';
import { rs } from '@rstest/core';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { showRemoveGoalModal } from '@monetr/interface/modals/RemoveGoalModal';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import Goals from '@monetr/interface/pages/goals';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('remove goal modal', () => {
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

  it('will remove the goal', async () => {
    apiSampleResponses(mockFetch);
    mockFetch
      .onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01gk23r0mgcekdf8bzkvf55f26')
      .reply(200);

    // Removing the goal prunes it from the cached spending list, so the list needs to be loaded first.
    const world = testRenderer(<Goals />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/goals' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveGoalModal({
          spending: new Spending({
            spendingId: ID.from<Spending>('spnd_01gk23r0mgcekdf8bzkvf55f26'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            name: 'Rainy Day',
            spendingType: SpendingType.Goal,
            targetAmount: 100000,
            currentAmount: 25000,
            usedAmount: 0,
            ruleset: null,
            lastRecurrence: null,
            nextRecurrence: '2099-04-15T05:00:00Z',
            nextContributionAmount: 5000,
            isBehind: false,
            isPaused: false,
            autoCreateTransaction: false,
            createdAt: '2026-03-16T06:00:00Z',
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-goal-modal')).toBeVisible());
    // Nothing should be sent until the user confirms
    expect(mockFetch.history.delete).toHaveLength(0);

    await act(() => user.click(world.getByTestId('remove-goal-confirm')));

    await waitFor(() => expect(world.queryByTestId('remove-goal-modal')).not.toBeInTheDocument());
    await waitFor(() => expect(onRemoved).toHaveBeenCalled());
    expect(mockFetch.history.delete).toHaveLength(1);
    expect(mockFetch.history.delete?.[0]).toMatchObject({
      url: '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01gk23r0mgcekdf8bzkvf55f26',
    });
    expect(mockEnqueueSnackbar).not.toHaveBeenCalled();
  });

  it('will not remove the goal when cancelled', async () => {
    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/goals' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveGoalModal({
          spending: new Spending({
            spendingId: ID.from<Spending>('spnd_01gk23r0mgcekdf8bzkvf55f26'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            name: 'Rainy Day',
            spendingType: SpendingType.Goal,
            targetAmount: 100000,
            currentAmount: 25000,
            usedAmount: 0,
            ruleset: null,
            lastRecurrence: null,
            nextRecurrence: '2099-04-15T05:00:00Z',
            nextContributionAmount: 5000,
            isBehind: false,
            isPaused: false,
            autoCreateTransaction: false,
            createdAt: '2026-03-16T06:00:00Z',
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-goal-modal')).toBeVisible());

    await user.click(world.getByTestId('close-remove-goal-modal'));

    // Make sure it goes away without sending anything.
    await waitFor(() => expect(world.queryByTestId('remove-goal-modal')).not.toBeInTheDocument());
    expect(mockFetch.history.delete).toHaveLength(0);
    expect(onRemoved).not.toHaveBeenCalled();
  });

  it('will show an error when the goal cannot be removed', async () => {
    mockFetch
      .onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending/spnd_01gk23r0mgcekdf8bzkvf55f26')
      .reply(400, {
        error: 'Failed to remove spending',
      });

    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/goals' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveGoalModal({
          spending: new Spending({
            spendingId: ID.from<Spending>('spnd_01gk23r0mgcekdf8bzkvf55f26'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            name: 'Rainy Day',
            spendingType: SpendingType.Goal,
            targetAmount: 100000,
            currentAmount: 25000,
            usedAmount: 0,
            ruleset: null,
            lastRecurrence: null,
            nextRecurrence: '2099-04-15T05:00:00Z',
            nextContributionAmount: 5000,
            isBehind: false,
            isPaused: false,
            autoCreateTransaction: false,
            createdAt: '2026-03-16T06:00:00Z',
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-goal-modal')).toBeVisible());

    await act(() => user.click(world.getByTestId('remove-goal-confirm')));

    // The modal stays open so the user can try again
    await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Failed to remove spending', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
    expect(world.getByTestId('remove-goal-modal')).toBeVisible();
    await waitFor(() => expect(world.getByTestId('remove-goal-confirm')).toBeEnabled());
    expect(onRemoved).not.toHaveBeenCalled();
  });
});
