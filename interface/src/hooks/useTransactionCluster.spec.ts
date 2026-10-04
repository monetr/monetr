import { waitFor } from '@testing-library/react';

import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import { ID } from '@monetr/interface/models/ID';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('read individual transaction clusters', () => {
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

  it('will request a single transaction cluster', async () => {
    mockFetch
      .onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/similar/tcl_01hy4rf0p7mz9w2q3c4v5b6n7m')
      .reply(200, {
        transactionClusterId: 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m',
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        name: 'Netflix',
        originalName: 'NETFLIX.COM',
        createdAt: '2026-03-15T06:00:00Z',
        updatedAt: '2026-03-15T06:00:00Z',
      });

    const world = testRenderHook(
      () => useTransactionCluster(ID.from<TransactionCluster>('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m')),
      {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
      },
    );
    await waitFor(() => expect(world.result.current.isLoading).toBeTruthy());
    await waitFor(() => expect(world.result.current.data).toBeDefined());
    expect(world.result.current.data?.transactionClusterId).toBe('tcl_01hy4rf0p7mz9w2q3c4v5b6n7m');
    expect(world.result.current.data?.name).toBe('Netflix');
  });

  it('will not request anything without an id', async () => {
    const world = testRenderHook(() => useTransactionCluster(null), {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/transactions',
    });
    await waitFor(() => expect(world.result.current.fetchStatus).toBe('idle'));
    expect(world.result.current.data).toBeUndefined();
  });
});
