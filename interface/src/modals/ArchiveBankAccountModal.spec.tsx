import { act } from 'react';
import { rs } from '@rstest/core';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import SelectBankAccount from '@monetr/interface/components/Layout/SelectBankAccount';
import { showArchiveBankAccountModal } from '@monetr/interface/modals/ArchiveBankAccountModal';
import BankAccount, { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import type Link from '@monetr/interface/models/Link';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';
import * as notifyActual from '@monetr/notify' with { rstest: 'importActual' };

const mockEnqueueSnackbar = rs.fn();
rs.mock('@monetr/notify', () => ({
  ...notifyActual,
  useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

describe('archive bank account modal', () => {
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

  it('will archive the bank account', async () => {
    apiSampleResponses(mockFetch);
    mockFetch.onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200);

    // Archiving the bank account prunes it from the cached bank account list, so the list needs to be loaded first.
    const world = testRenderer(<SelectBankAccount />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/settings',
    });
    const user = userEvent.setup();
    const onArchived = rs.fn();
    await waitFor(() => expect(world.getByText('Mercury Checking')).toBeVisible());

    // Open the dialog
    await act(
      () =>
        void showArchiveBankAccountModal({
          bankAccount: new BankAccount({
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            linkId: ID.from<Link>('link_01gds6eqsqacg48p0azb3wcpsq'),
            lunchFlowBankAccountId: null,
            mask: '2982',
            name: 'Mercury Checking',
            originalName: 'Mercury Checking',
            status: 'active',
            accountType: BankAccountType.Depository,
            accountSubType: BankAccountSubType.Checking,
            currency: 'USD',
            currentBalance: 47986,
            availableBalance: 47986,
            limitBalance: null,
            lastUpdated: '2024-08-27T08:53:48.555368Z',
            createdAt: '2022-09-25T02:08:40.758642Z',
            createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
            deletedAt: null,
            plaidBankAccount: null,
            lunchFlowBankAccount: null,
          }),
        }).then(onArchived),
    );
    await waitFor(() => expect(world.getByTestId('archive-bank-account-modal')).toBeVisible());
    // Nothing should be sent until the user confirms
    expect(mockFetch.history.delete).toHaveLength(0);

    await act(() => user.click(world.getByTestId('archive-bank-account-confirm')));

    await waitFor(() => expect(world.queryByTestId('archive-bank-account-modal')).not.toBeInTheDocument());
    await waitFor(() => expect(onArchived).toHaveBeenCalled());
    expect(mockFetch.history.delete).toHaveLength(1);
    expect(mockFetch.history.delete?.[0]).toMatchObject({
      url: '/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb',
    });
    expect(mockEnqueueSnackbar).not.toHaveBeenCalled();
  });

  it('will not archive the bank account when cancelled', async () => {
    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/settings' });
    const user = userEvent.setup();
    const onArchived = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showArchiveBankAccountModal({
          bankAccount: new BankAccount({
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            linkId: ID.from<Link>('link_01gds6eqsqacg48p0azb3wcpsq'),
            lunchFlowBankAccountId: null,
            mask: '2982',
            name: 'Mercury Checking',
            originalName: 'Mercury Checking',
            status: 'active',
            accountType: BankAccountType.Depository,
            accountSubType: BankAccountSubType.Checking,
            currency: 'USD',
            currentBalance: 47986,
            availableBalance: 47986,
            limitBalance: null,
            lastUpdated: '2024-08-27T08:53:48.555368Z',
            createdAt: '2022-09-25T02:08:40.758642Z',
            createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
            deletedAt: null,
            plaidBankAccount: null,
            lunchFlowBankAccount: null,
          }),
        }).then(onArchived),
    );
    await waitFor(() => expect(world.getByTestId('archive-bank-account-modal')).toBeVisible());

    await user.click(world.getByTestId('close-archive-bank-account-modal'));

    // Make sure it goes away without sending anything.
    await waitFor(() => expect(world.queryByTestId('archive-bank-account-modal')).not.toBeInTheDocument());
    expect(mockFetch.history.delete).toHaveLength(0);
    expect(onArchived).not.toHaveBeenCalled();
  });

  it('will show an error when the bank account cannot be archived', async () => {
    mockFetch.onDelete('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(400, {
      error: 'Failed to archive bank account',
    });

    const world = testRenderer(<div />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/settings' });
    const user = userEvent.setup();
    const onArchived = rs.fn();

    // Open the dialog
    await act(
      () =>
        void showArchiveBankAccountModal({
          bankAccount: new BankAccount({
            bankAccountId: ID.from<BankAccount>('bac_01gds6eqsq7h5mgevwtmw3cyxb'),
            linkId: ID.from<Link>('link_01gds6eqsqacg48p0azb3wcpsq'),
            lunchFlowBankAccountId: null,
            mask: '2982',
            name: 'Mercury Checking',
            originalName: 'Mercury Checking',
            status: 'active',
            accountType: BankAccountType.Depository,
            accountSubType: BankAccountSubType.Checking,
            currency: 'USD',
            currentBalance: 47986,
            availableBalance: 47986,
            limitBalance: null,
            lastUpdated: '2024-08-27T08:53:48.555368Z',
            createdAt: '2022-09-25T02:08:40.758642Z',
            createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
            deletedAt: null,
            plaidBankAccount: null,
            lunchFlowBankAccount: null,
          }),
        }).then(onArchived),
    );
    await waitFor(() => expect(world.getByTestId('archive-bank-account-modal')).toBeVisible());

    await act(() => user.click(world.getByTestId('archive-bank-account-confirm')));

    // The modal stays open so the user can try again
    await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledTimes(1));
    expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Failed to archive bank account', {
      variant: 'error',
      disableWindowBlurListener: true,
    });
    expect(world.getByTestId('archive-bank-account-modal')).toBeVisible();
    await waitFor(() => expect(world.getByTestId('archive-bank-account-confirm')).toBeEnabled());
    expect(onArchived).not.toHaveBeenCalled();
  });
});
