import React, { act, Fragment } from 'react';
import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { showNewBankAccountModal } from '@monetr/interface/modals/NewBankAccountModal';
import testRenderer from '@monetr/interface/testutils/renderer';

const mockUseNavigate = jest.fn((_url: string) => {});
jest.mock('react-router-dom', () => ({
  __esModule: true,
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockUseNavigate,
}));

describe('new bank account modal', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
    mockUseNavigate.mockReset();
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will render', async () => {
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'linkId': 'link_01gds6eqsqacg48p0azb3wcpsq',
      'availableBalance': 47986,
      'currentBalance': 47986,
      'mask': '2982',
      'name': 'Mercury Checking',
      'originalName': 'Mercury Checking',
      'accountType': 'depository',
      'accountSubType': 'checking',
      'status': 'active',
      'lastUpdated': '2024-08-27T08:53:48.555368Z',
      'createdAt': '2022-09-25T02:08:40.758642Z',
      'updatedAt': '2024-03-19T06:17:32.335106Z',
    });
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules').reply(200, []);

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions' });
    // Open the dialog
    await act(() => void showNewBankAccountModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-bank-account-modal')).toBeVisible());
    // Close the dialog.
    act(() => world.getByTestId('close-new-bank-account-modal').click());
    // Make sure it goes away.
    await waitFor(() => expect(world.queryByTestId('new-bank-account-modal')).not.toBeInTheDocument());
  });

  it('will attempt to create a new bank account', async () => {
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'linkId': 'link_01gds6eqsqacg48p0azb3wcpsq',
      'availableBalance': 47986,
      'currentBalance': 47986,
      'mask': '2982',
      'name': 'Mercury Checking',
      'originalName': 'Mercury Checking',
      'accountType': 'depository',
      'accountSubType': 'checking',
      'status': 'active',
      'lastUpdated': '2024-08-27T08:53:48.555368Z',
      'createdAt': '2022-09-25T02:08:40.758642Z',
      'updatedAt': '2024-03-19T06:17:32.335106Z',
    });
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules').reply(200, []);

    mockAxios.onPost('/api/bank_accounts').reply(200, {
      'bankAccountId': 'bac_created',
      'linkId': 'link_01gds6eqsqacg48p0azb3wcpsq',
      'availableBalance': 10000,
      'currentBalance': 10000,
      'mask': '',
      'name': 'Test Account',
      'originalName': 'Test Account',
      'accountType': 'depository',
      'accountSubType': 'checking',
      'status': 'active',
    });

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions' });
    // Open the dialog
    await act(() => void showNewBankAccountModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-bank-account-modal')).toBeVisible());

    // Fill out the modal's form and submit it.
    await act(() => userEvent.type(world.getByTestId('bank-account-name'), 'Test Account'));
    await act(() => userEvent.type(world.getByTestId('bank-account-balance'), '100'));
    await act(() => userEvent.click(world.getByTestId('bank-account-submit')));

    // When we submit it we should get redirected to our new bank account.
    await waitFor(() => expect(mockUseNavigate).toBeCalledWith('/bank/bac_created/transactions'));

    // Make sure the modal was also closed.
    await waitFor(() => expect(world.queryByTestId('new-bank-account-modal')).not.toBeInTheDocument());
  });
});
