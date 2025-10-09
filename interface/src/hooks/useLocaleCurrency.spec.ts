import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('use locale currency', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will provide defaults if youre not authenticated', async () => {
    mockAxios.onGet('/users/me').reply(403, {
      error: 'unauthenticated',
    });

    const { result, waitFor } = testRenderHook(useLocaleCurrency, { initialRoute: '/login' });
    expect(result.current.isFetching).toBeTruthy();
    await waitFor(() => expect(result.current.isFetching).toBeFalsy());
    await waitFor(() => expect(result.current.isLoading).toBeFalsy());
    await waitFor(() => expect(result.current.data.locale).toBe('en_US'));
    await waitFor(() => expect(result.current.data.currency).toBe('USD'));
  });

  it('will provide the real locale and currency for the current user', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
      'activeUntil': '2024-09-26T00:31:38Z',
      'hasSubscription': true,
      'isActive': true,
      'isSetup': true,
      'isTrialing': false,
      'trialingUntil': null,
      'defaultCurrency': 'JPY',
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
          'locale': 'ja_JP',
          'subscriptionActiveUntil': '2024-09-26T00:31:38Z',
          'subscriptionStatus': 'active',
          'trialEndsAt': null,
          'createdAt': '2024-01-03T17:02:23.290914Z',
        },
      },
    });

    const { result, waitFor } = testRenderHook(useLocaleCurrency, { initialRoute: '/setup' });
    expect(result.current.isFetching).toBeTruthy();
    await waitFor(() => expect(result.current.isFetching).toBeFalsy());
    await waitFor(() => expect(result.current.isLoading).toBeFalsy());
    await waitFor(() => expect(result.current.data.locale).toBe('ja_JP'));
    await waitFor(() => expect(result.current.data.currency).toBe('JPY'));
  });

  it('will provide the locale for the current bank account if it can', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
      'activeUntil': '2024-09-26T00:31:38Z',
      'hasSubscription': true,
      'isActive': true,
      'isSetup': true,
      'isTrialing': false,
      'trialingUntil': null,
      'defaultCurrency': 'JPY',
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
          'locale': 'ja_JP',
          'subscriptionActiveUntil': '2024-09-26T00:31:38Z',
          'subscriptionStatus': 'active',
          'trialEndsAt': null,
          'createdAt': '2024-01-03T17:02:23.290914Z',
        },
      },
    });
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'linkId': 'link_01gds6eqsqacg48p0azb3wcpsq',
      'availableBalance': 47986,
      'currentBalance': 47986,
      'mask': '2982',
      'name': 'Mercury Checking',
      'originalName': 'Mercury Checking',
      'accountType': 'depository',
      'accountSubType': 'checking',
      'status': 'active',
      'currency': 'EUR',
      'lastUpdated': '2024-08-27T08:53:48.555368Z',
      'createdAt': '2022-09-25T02:08:40.758642Z',
      'updatedAt': '2024-03-19T06:17:32.335106Z',
    });

    const { result, waitFor } = testRenderHook(useLocaleCurrency, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions',
    });
    expect(result.current.isFetching).toBeTruthy();
    await waitFor(() => expect(result.current.isFetching).toBeFalsy());
    await waitFor(() => expect(result.current.isLoading).toBeFalsy());
    // Locale should always come from the user.
    await waitFor(() => expect(result.current.data.locale).toBe('ja_JP'));
    // But currency should come from the bank account when there is one.
    await waitFor(() => expect(result.current.data.currency).toBe('EUR'));
  });

  it('will handle a bad bank ID', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
      'activeUntil': '2024-09-26T00:31:38Z',
      'hasSubscription': true,
      'isActive': true,
      'isSetup': true,
      'isTrialing': false,
      'trialingUntil': null,
      'defaultCurrency': 'JPY',
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
          'locale': 'ja_JP',
          'subscriptionActiveUntil': '2024-09-26T00:31:38Z',
          'subscriptionStatus': 'active',
          'trialEndsAt': null,
          'createdAt': '2024-01-03T17:02:23.290914Z',
        },
      },
    });
    mockAxios.onGet('/api/bank_accounts/undefined').reply(404, {
      'error': 'Not found',
    });

    const { result, waitFor } = testRenderHook(useLocaleCurrency, { initialRoute: '/bank/undefined/transactions' });
    expect(result.current.isFetching).toBeTruthy();
    await waitFor(() => expect(result.current.isFetching).toBeFalsy());
    // Still use the user's locale
    await waitFor(() => expect(result.current.data.locale).toBe('en_US'));
    // But fall back to the default if we can't load the bank currency
    await waitFor(() => expect(result.current.data.currency).toBe('USD'));
  });
});
