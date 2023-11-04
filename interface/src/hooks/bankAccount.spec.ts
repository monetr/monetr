import { rest } from 'msw';

import { useSelectedBankAccount, useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import testRenderHook from '@monetr/interface/testutils/hooks';
import { server } from '@monetr/interface/testutils/server';

describe('bank account hooks', () => {
  describe('useSelectedBankAccount', () => {
    it('valid URL', async () => {
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

      { // Make sure use selected bank account works.
        const world = testRenderHook(useSelectedBankAccount, { initialRoute: '/bank/12/expenses' });
        expect(world.result.current.data).not.toBeDefined();
        expect(world.result.current.isLoading).toBeTruthy();
        await world.waitFor(() => expect(world.result.current.isSuccess).toBeTruthy());
        expect(world.result.current.data.bankAccountId).toBe(12);
      }

      { // Then make sure that useSelectedBankAccountId works
        const world = testRenderHook(useSelectedBankAccountId, { initialRoute: '/bank/12/expenses' });
        expect(world.result.current).toBeUndefined();
        await world.waitFor(() => expect(world.result.current).toBeDefined());
        expect(world.result.current).toBe(12);
      }
    });

    it('invalid url', async () => {
      { // useSelectedBankAccount
        const { result } = testRenderHook(useSelectedBankAccount, { initialRoute: '/bank/bad/expenses' });
        expect(result.error).toBeDefined();
        expect(result.error.message).toBe('invalid bank account ID specified: "bad" is not a valid bank account ID');
      }

      { // useSelectedBankAccountId
        const { result } = testRenderHook(useSelectedBankAccountId, { initialRoute: '/bank/bad/expenses' });
        expect(result.error).toBeDefined();
        expect(result.error.message).toBe('invalid bank account ID specified: "bad" is not a valid bank account ID');
      }
    });

    it('non-bank url selected bank account basic', async () => {
      const { result } = testRenderHook(useSelectedBankAccount, { initialRoute: '/settings' });
      expect(result.error).toBeUndefined();
      // When we are not _enabled_, we will always have is loading set to true.
      expect(result.current.isLoading).toBeTruthy();
      // But we won't be fetching!
      expect(result.current.isFetching).toBeFalsy();
      // Because of the URL, the current bank account should be null.
      expect(result.current.data).toBeUndefined();
    });

    it('non-bank url selected bank account ID', async () => {
      const { result } = testRenderHook(useSelectedBankAccountId, { initialRoute: '/settings' });
      expect(result.error).toBeUndefined();
      // Because of the URL, the current bank account ID should be null.
      expect(result.current).toBeUndefined();
    });
  });
});
