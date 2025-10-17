import React from 'react';
import { waitFor } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import Monetr from '@monetr/interface/monetr';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('monetr app', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will render the demo app used in docs', async () => {
    apiSampleResponses(mockAxios);

    const world = testRenderer(<Monetr />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions',
    });

    await waitFor(() => expect(world.getByTestId('bank-sidebar-subscription')).toBeVisible());
    // Make sure that at least one of our transactions renders.
    await waitFor(() => expect(world.getByTestId('txn_01j68vszqeq30t7jz7atk9yd9r')).toBeVisible());
  });
});
