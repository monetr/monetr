import React, { Fragment } from 'react';
import { act, waitFor } from '@testing-library/react';
import { rest } from 'msw';

import { showNewFundingModal } from '@monetr/interface/modals/NewFundingModal';
import testRenderer from '@monetr/interface/testutils/renderer';
import { server } from '@monetr/interface/testutils/server';

describe('new funding schedule modal', () => {
  it('will render', async () => {
    server.use(
      rest.get('/api/bank_accounts/12', (_req, res, ctx) => {
        return res(ctx.json({
          'bankAccountId': 12,
          'linkId': 4,
          'availableBalance': 48635,
          'currentBalance': 48635,
          'mask': '2982',
          'name': 'Mercury Checking',
          'originalName': 'Mercury Checking',
          'officialName': 'Mercury Checking',
          'accountType': 'depository',
          'accountSubType': 'checking',
          'status': 'active',
          'lastUpdated': '2023-07-02T04:22:52.48118Z',
        }));
      }),
    );

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/12/funding' });
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
