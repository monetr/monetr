import { waitFor } from '@testing-library/react';

import { useTransactionClusterTransactions } from '@monetr/interface/hooks/useTransactionClusterTransactions';
import { ID } from '@monetr/interface/models/ID';
import type TransactionCluster from '@monetr/interface/models/TransactionCluster';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

const bankAccountId = 'bac_01hy4rcmadc01d2kzv7vynbxxx';
const clusterId = 'tcl_01hy4rf0p7mz9w2q3c4v5b6n7m';

describe('transaction cluster transactions', () => {
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

  it('will request the most recent 100 in the cluster', async () => {
    mockFetch
      .onGet(`/api/bank_accounts/${bankAccountId}/transactions?transaction_cluster_id=${clusterId}&limit=100`)
      .reply(200, [
        {
          transactionId: 'txn_01hy4rh2k8m3n4p5q6r7s8t9v0',
          bankAccountId,
          transactionClusterId: clusterId,
          amount: 800,
          date: '2026-03-15T05:00:00Z',
          name: 'Github',
          originalName: 'GITHUB.COM 877-448-4820 CA',
          isPending: false,
          createdAt: '2026-03-15T06:00:00Z',
        },
      ]);

    const world = testRenderHook(() => useTransactionClusterTransactions(ID.from<TransactionCluster>(clusterId)), {
      initialRoute: `/bank/${bankAccountId}/recurring`,
    });
    await waitFor(() => expect(world.result.current.data).toHaveLength(1));
    expect(world.result.current.data?.[0]?.originalName).toBe('GITHUB.COM 877-448-4820 CA');
    expect(world.result.current.data?.[0]?.date).toBeInstanceOf(Date);
  });

  it('will not request anything without a cluster', async () => {
    const world = testRenderHook(() => useTransactionClusterTransactions(null), {
      initialRoute: `/bank/${bankAccountId}/recurring`,
    });
    await waitFor(() => expect(world.result.current.fetchStatus).toBe('idle'));
    expect(world.result.current.data).toBeUndefined();
  });
});
