import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('use the selected bank account ID', () => {
  it('will parse the URL properly', () => {
    const world = testRenderHook(useSelectedBankAccountId, {
      initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/expenses',
    });
    expect(world.result.current).toBe('bac_01hy4rcmadc01d2kzv7vynbxxx');
  });

  it('will return undefined on an invalid url', () => {
    const world = testRenderHook(useSelectedBankAccountId, {
      initialRoute: '/settings',
    });
    expect(world.result.current).toBeUndefined();
  });
});
