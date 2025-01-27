import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { useInstalledCurrencies } from '@monetr/interface/hooks/useInstalledCurrencies';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('use installed currencies', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will fetch currencies if we are authenticated', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
      'activeUntil': '2024-09-26T00:31:38Z',
      'hasSubscription': true,
      'isActive': true,
      'isSetup': true,
      'isTrialing': false,
      'trialingUntil': null,
      'defaultCurrency': 'USD',
      'user': {
        'userId': 'user_01hym36e8ewaq0hxssb1m3k4ha',
        'loginId': 'lgn_01hym36d96ze86vz5g7883vcwg',
        'login': {
          'loginId': 'lgn_01hym36d96ze86vz5g7883vcwg',
          'email': 'example@example.com',
          'firstName': 'Elliot',
          'lastName': 'Courant',
          'passwordResetAt': null,
          'isEmailVerified': true,
          'emailVerifiedAt': '2022-09-25T00:24:25.976514Z',
          'totpEnabledAt': null,
        },
        'accountId': 'acct_01hk84dchvxvjgp7cgap818c82',
        'account': {
          'accountId': 'acct_01hk84dchvxvjgp7cgap818c82',
          'timezone': 'America/Chicago',
          'locale': 'en_US',
          'subscriptionActiveUntil': '2024-09-26T00:31:38Z',
          'subscriptionStatus': 'active',
          'trialEndsAt': null,
          'createdAt': '2024-01-03T17:02:23.290914Z',
        },
      },
    });
    mockAxios.onGet('/api/locale/currency').reply(200, [
      'EUR',
      'USD',
      // Having all of them doesn't matter, just testing
    ]);

    const world = testRenderHook(useInstalledCurrencies, {
      initialRoute: '/settings',
    });
    expect(world.result.current.data).not.toBeDefined();
    expect(world.result.current.isLoading).toBeTruthy();
    await world.waitFor(() => expect(world.result.current.isFetching).toBeTruthy());
    await world.waitFor(() => expect(world.result.current.isSuccess).toBeTruthy());
    expect(world.result.current.data).toStrictEqual([
      'EUR',
      'USD',
    ]);
  });

  it('will not fetch currencies if we are not authenticated', async () => {
    mockAxios.onGet('/users/me').reply(403, {
      error: 'unauthenticated',
    });
    const world = testRenderHook(useInstalledCurrencies, {
      initialRoute: '/login',
    });
    expect(world.result.current.data).not.toBeDefined();
    expect(world.result.current.isLoading).toBeTruthy();
    // Stays falsy
    expect(world.result.current.isFetching).toBeFalsy();
  });
});
