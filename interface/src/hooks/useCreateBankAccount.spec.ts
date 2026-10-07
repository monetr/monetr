import { act } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { useCreateBankAccount } from '@monetr/interface/hooks/useCreateBankAccount';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('create bank account', () => {
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

  it('will leave the bank account list alone when it was never loaded', async () => {
    mockFetch.onPost('/api/bank_accounts').reply(200, {
      bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      availableBalance: 48635,
      currentBalance: 48635,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      currency: 'USD',
      lastUpdated: '2026-03-16T06:00:00Z',
      createdAt: '2026-03-16T06:00:00Z',
      updatedAt: '2026-03-16T06:00:00Z',
    });

    const world = testRenderHook(
      () => ({
        createBankAccount: useCreateBankAccount(),
        queryClient: useQueryClient(),
      }),
      {
        initialRoute: '/setup',
      },
    );
    await act(async () => {
      await world.result.current.createBankAccount({
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
        name: 'Mercury Checking',
        mask: '2982',
        availableBalance: 48635,
        currentBalance: 48635,
        accountType: BankAccountType.Depository,
        accountSubType: BankAccountSubType.Checking,
        currency: 'USD',
      });
    });

    // Caching a list with just the new one in it would hide every other bank account until something refetched it
    expect(world.result.current.queryClient.getQueryData(['GET', '/api/bank_accounts'])).toBeUndefined();
    expect(
      world.result.current.queryClient.getQueryData(['GET', '/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx']),
    ).toBeDefined();
  });
});
