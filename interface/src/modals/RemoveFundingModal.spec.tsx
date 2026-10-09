import { act } from 'react';
import { rs } from '@rstest/core';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { showRemoveFundingModal } from '@monetr/interface/modals/RemoveFundingModal';
import type BankAccount from '@monetr/interface/models/BankAccount';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { ID } from '@monetr/interface/models/ID';
import Funding from '@monetr/interface/pages/funding';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('remove funding modal', () => {
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

  it('will remove the funding schedule', async () => {
    apiSampleResponses(mockFetch);
    mockFetch
      .onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules/fund_01hym37k3kj4ghv67nfx7vkvr0')
      .reply(200);

    // Removing the funding schedule prunes it from the cached funding list, so the list needs to be loaded first.
    const world = testRenderer(<Funding />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();
    await waitFor(() => expect(world.getByText('Funding Schedules')).toBeVisible());

    // Open the dialog
    await act(
      () =>
        void showRemoveFundingModal({
          funding: new FundingSchedule({
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            name: "Elliot's Contribution",
            description: '15th and last day of every month',
            ruleset: 'DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
            lastRecurrence: '2024-08-15T05:00:00Z',
            nextRecurrence: '2024-08-30T05:00:00Z',
            nextRecurrenceOriginal: '2024-08-31T05:00:00Z',
            excludeWeekends: true,
            autoCreateTransaction: false,
            estimatedDeposit: 22000,
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-funding-modal')).toBeVisible());
    // Nothing should be sent until the user confirms
    expect(mockFetch.history.delete).toHaveLength(0);

    await act(() => user.click(world.getByTestId('remove-funding-confirm')));

    await waitFor(() => expect(world.queryByTestId('remove-funding-modal')).not.toBeInTheDocument());
    await waitFor(() => expect(onRemoved).toHaveBeenCalled());
    expect(mockFetch.history.delete).toHaveLength(1);
    expect(mockFetch.history.delete?.[0]).toMatchObject({
      url: '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules/fund_01hym37k3kj4ghv67nfx7vkvr0',
    });
    expect(mockEnqueueSnackbar).not.toHaveBeenCalled();
  });

  it('will not remove the funding schedule when cancelled', async () => {
    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveFundingModal({
          funding: new FundingSchedule({
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            name: "Elliot's Contribution",
            description: '15th and last day of every month',
            ruleset: 'DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
            lastRecurrence: '2024-08-15T05:00:00Z',
            nextRecurrence: '2024-08-30T05:00:00Z',
            nextRecurrenceOriginal: '2024-08-31T05:00:00Z',
            excludeWeekends: true,
            autoCreateTransaction: false,
            estimatedDeposit: 22000,
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-funding-modal')).toBeVisible());

    await user.click(world.getByTestId('close-remove-funding-modal'));

    // Make sure it goes away without sending anything.
    await waitFor(() => expect(world.queryByTestId('remove-funding-modal')).not.toBeInTheDocument());
    expect(mockFetch.history.delete).toHaveLength(0);
    expect(onRemoved).not.toHaveBeenCalled();
  });

  it('will show an error when the funding schedule cannot be removed', async () => {
    mockFetch
      .onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules/fund_01hym37k3kj4ghv67nfx7vkvr0')
      .reply(400, {
        error: 'Failed to remove funding schedule',
      });

    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding' });
    const user = userEvent.setup();
    const onRemoved = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showRemoveFundingModal({
          funding: new FundingSchedule({
            fundingScheduleId: ID.from<FundingSchedule>('fund_01hym37k3kj4ghv67nfx7vkvr0'),
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            name: "Elliot's Contribution",
            description: '15th and last day of every month',
            ruleset: 'DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
            lastRecurrence: '2024-08-15T05:00:00Z',
            nextRecurrence: '2024-08-30T05:00:00Z',
            nextRecurrenceOriginal: '2024-08-31T05:00:00Z',
            excludeWeekends: true,
            autoCreateTransaction: false,
            estimatedDeposit: 22000,
          }),
        }).then(onRemoved),
    );
    await waitFor(() => expect(world.getByTestId('remove-funding-modal')).toBeVisible());

    await act(() => user.click(world.getByTestId('remove-funding-confirm')));

    // The modal stays open so the user can try again
    await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Failed to remove funding schedule', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
    expect(world.getByTestId('remove-funding-modal')).toBeVisible();
    await waitFor(() => expect(world.getByTestId('remove-funding-confirm')).toBeEnabled());
    expect(onRemoved).not.toHaveBeenCalled();
  });
});
