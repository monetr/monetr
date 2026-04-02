import { waitFor } from '@testing-library/react';

import Monetr from '@monetr/interface/monetr';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('monetr app', () => {
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

  it('will render the demo app used in docs', async () => {
    apiSampleResponses(mockFetch);

    const world = testRenderer(<Monetr />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions',
    });

    await waitFor(() => expect(world.getByTestId('bank-sidebar-subscription')).toBeVisible());
    // Make sure that at least one of our transactions renders.
    await waitFor(() => expect(world.getByTestId('txn_01j68vszqeq30t7jz7atk9yd9r')).toBeVisible());
  });
});
