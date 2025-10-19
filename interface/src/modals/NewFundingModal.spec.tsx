import { Fragment } from 'react';

import { act, waitFor } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('new funding schedule modal', () => {
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

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding' });
    // Open the dialog
    await act(() => void showNewFundingModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-funding-modal')).toBeVisible());
    // Close the dialog.
    act(() => world.getByTestId('close-new-funding-modal').click());
    // Make sure it goes away.
    await waitFor(() => expect(world.queryByTestId('new-funding-modal')).not.toBeInTheDocument());
  });
});
