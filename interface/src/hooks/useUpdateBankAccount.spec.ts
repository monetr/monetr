import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { useUpdateBankAccount } from '@monetr/interface/hooks/useUpdateBankAccount';
import BankAccount from '@monetr/interface/models/BankAccount';
import { ID } from '@monetr/interface/models/ID';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('update bank account', () => {
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

  it('will update the bank account', async () => {
    mockFetch.onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      mask: '2982',
      name: 'Renamed Checking',
      originalName: 'Mercury Checking',
      status: 'active',
      accountType: 'depository',
      accountSubType: 'checking',
      currency: 'USD',
      availableBalance: 50000,
      currentBalance: 50000,
      limitBalance: null,
      lastUpdated: '2023-07-02T04:22:52.48118Z',
    });

    // We render the query client alongside the hook so we can poke at the cache. In the real app the bank accounts list
    // is always loaded before you ever reach the settings page, but in a test we have to seed it ourselves otherwise
    // the onSuccess previous.map blows up on undefined.
    const world = testRenderHook(
      () => ({
        updateBankAccount: useUpdateBankAccount(),
        queryClient: useQueryClient(),
      }),
      { initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/settings' },
    );

    act(() => {
      world.result.current.queryClient.setQueryData(
        ['/api/bank_accounts'],
        [
          {
            bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
            name: 'Mercury Checking',
          },
        ],
      );
    });

    let result!: BankAccount;
    await act(async () => {
      result = await world.result.current.updateBankAccount({
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Renamed Checking',
        currency: 'USD',
        availableBalance: 50000,
        currentBalance: 50000,
        limitBalance: null,
      });
    });

    // The hook should hand us back a real BankAccount instance, not the raw json, otherwise the getters everything
    // downstream relies on wont exist.
    expect(result).toBeInstanceOf(BankAccount);
    expect(result.name).toBe('Renamed Checking');

    // Make sure we actually PATCHed and did not fall back to the old PUT route.
    const patchHistory = mockFetch.history.patch;
    expect(patchHistory).toHaveLength(1);
    expect(patchHistory?.[0]?.url).toBe('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx');

    // The single bank account cache should be set directly to the updated account so anything reading it gets the new
    // values without a refetch.
    const single = world.result.current.queryClient.getQueryData<BankAccount>([
      '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx',
    ]);
    expect(single).toBeInstanceOf(BankAccount);
    expect(single?.name).toBe('Renamed Checking');

    // And the entry in the list cache should have been swapped out for the updated account too.
    const list = world.result.current.queryClient.getQueryData<Array<BankAccount>>(['/api/bank_accounts']);
    expect(list?.[0]?.name).toBe('Renamed Checking');
  });

  it('it will fail to update the bank account', async () => {
    mockFetch.onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(400, {
      error: 'Invalid request',
    });

    const world = testRenderHook(
      () => ({
        updateBankAccount: useUpdateBankAccount(),
        queryClient: useQueryClient(),
      }),
      { initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/settings' },
    );

    await act(async () => {
      await expect(
        world.result.current.updateBankAccount({
          bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
          name: 'Renamed Checking',
        }),
      ).rejects.toMatchObject({
        message: 'Request failed with status code 400',
        response: {
          data: {
            error: 'Invalid request',
          },
        },
      });
    });
  });

  it('will drop the bankAccountId from the request body but keep nulls', async () => {
    mockFetch.onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      name: 'Renamed Checking',
      originalName: 'Mercury Checking',
      status: 'active',
      accountType: 'depository',
      accountSubType: 'checking',
      currency: 'USD',
      availableBalance: 50000,
      currentBalance: 50000,
      limitBalance: null,
      lastUpdated: '2023-07-02T04:22:52.48118Z',
    });

    const world = testRenderHook(
      () => ({
        updateBankAccount: useUpdateBankAccount(),
        queryClient: useQueryClient(),
      }),
      { initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/settings' },
    );

    act(() => {
      world.result.current.queryClient.setQueryData(
        ['/api/bank_accounts'],
        [
          {
            bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
            name: 'Mercury Checking',
          },
        ],
      );
    });

    await act(async () => {
      await world.result.current.updateBankAccount({
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Renamed Checking',
        // limitBalance is explicitly null because the user is clearing the limit, so it MUST survive into the body. A
        // missing key and a null key mean different things to the patch endpoint.
        limitBalance: null,
      });
    });

    const body = mockFetch.history.patch?.[0]?.data as Record<string, unknown>;
    // bankAccountId is a path param, the hook destructures it out so it should never end up in the patch body.
    expect('bankAccountId' in body).toBe(false);
    expect('limitBalance' in body).toBe(true);
    expect(body.limitBalance).toBeNull();
    expect(body.name).toBe('Renamed Checking');
  });

  it('will invalidate the cached balances for the bank account', async () => {
    mockFetch.onPatch('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      name: 'Renamed Checking',
      originalName: 'Mercury Checking',
      status: 'active',
      accountType: 'depository',
      accountSubType: 'checking',
      currency: 'USD',
      availableBalance: 50000,
      currentBalance: 50000,
      limitBalance: null,
      lastUpdated: '2023-07-02T04:22:52.48118Z',
    });

    const world = testRenderHook(
      () => ({
        updateBankAccount: useUpdateBankAccount(),
        queryClient: useQueryClient(),
      }),
      { initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/settings' },
    );

    // Seed the list so onSuccess does not explode, and seed the balances query so we can prove it gets invalidated when
    // the balances change underneath it.
    act(() => {
      world.result.current.queryClient.setQueryData(
        ['/api/bank_accounts'],
        [
          {
            bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
            name: 'Mercury Checking',
          },
        ],
      );
      world.result.current.queryClient.setQueryData(['/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/balances'], {
        available: 50000,
        current: 50000,
      });
    });

    await act(async () => {
      await world.result.current.updateBankAccount({
        bankAccountId: ID.from<BankAccount>('bac_01hy4rcmadc01d2kzv7vynbxxx'),
        name: 'Renamed Checking',
        availableBalance: 60000,
        currentBalance: 60000,
      });
    });

    // After the update the balances query should be flagged stale so the next time something reads it the new numbers
    // get pulled down from the server.
    const state = world.result.current.queryClient.getQueryState([
      '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/balances',
    ]);
    expect(state?.isInvalidated).toBe(true);
  });
});
