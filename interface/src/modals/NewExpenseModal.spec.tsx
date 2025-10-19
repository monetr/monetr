import { act, Fragment } from 'react';

import { waitFor } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('new expense modal', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will render', async () => {
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
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
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules').reply(200, []);

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses' });
    // Open the dialog
    await act(() => void showNewExpenseModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-expense-modal')).toBeVisible());
    // Close the dialog.
    act(() => world.getByTestId('close-new-expense-modal').click());
    // Make sure it goes away.
    await waitFor(() => expect(world.queryByTestId('new-expense-modal')).not.toBeInTheDocument());
  });
});
