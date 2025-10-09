import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('useSelectedBankAccount', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('valid URL', async () => {
    mockAxios.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx').reply(200, {
      'bankAccountId': 'bac_01hy4rcmadc01d2kzv7vynbxxx', // 12,
      'linkId': 'link_01hy4rbb1gjdek7h2xmgy5pnwk', // 4
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
    });

    { // Make sure use selected bank account works.
      const world = testRenderHook(useSelectedBankAccount, {
        initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/expenses',
      });
      expect(world.result.current.data).not.toBeDefined();
      expect(world.result.current.isLoading).toBeTruthy();
      await world.waitFor(() => expect(world.result.current.isSuccess).toBeTruthy());
      expect(world.result.current.data.bankAccountId).toBe('bac_01hy4rcmadc01d2kzv7vynbxxx');
    }
  });

  it('non-bank url selected bank account basic', async () => {
    const { result } = testRenderHook(useSelectedBankAccount, {
      initialRoute: '/settings',
    });
    expect(result.error).toBeUndefined();
    expect(result.current.isLoading).toBeFalsy();
    expect(result.current.isFetching).toBeFalsy();
    // Because of the URL, the current bank account should be null.
    expect(result.current.data).toBeUndefined();
  });
});
