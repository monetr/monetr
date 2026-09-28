import { waitFor } from '@testing-library/react';

import { useCurrency } from '@monetr/interface/hooks/useCurrency';
import { useInstalledCurrencies } from '@monetr/interface/hooks/useInstalledCurrencies';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

const euro = {
  code: 'EUR',
  name: 'Euro',
  symbol: '€',
  decimalSeparator: '.',
  groupSeparator: ',',
  minusSign: '-',
  fractionalDigits: 2,
};

const yen = {
  code: 'JPY',
  name: 'Japanese Yen',
  symbol: '¥',
  decimalSeparator: '.',
  groupSeparator: ',',
  minusSign: '-',
  fractionalDigits: 0,
};

describe('use currency', () => {
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

  function givenIAmAuthenticated() {
    mockFetch.onGet('/api/users/me').reply(200, {
      activeUntil: '2024-09-26T00:31:38Z',
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      trialingUntil: null,
      defaultCurrency: 'USD',
      user: {
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        loginId: 'lgn_01hym36d96ze86vz5g7883vcwg',
        login: {
          loginId: 'lgn_01hym36d96ze86vz5g7883vcwg',
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          passwordResetAt: null,
          isEmailVerified: true,
          emailVerifiedAt: '2022-09-25T00:24:25.976514Z',
          totpEnabledAt: null,
        },
        accountId: 'acct_01hk84dchvxvjgp7cgap818c82',
        account: {
          accountId: 'acct_01hk84dchvxvjgp7cgap818c82',
          timezone: 'America/Chicago',
          locale: 'en_US',
          subscriptionActiveUntil: '2024-09-26T00:31:38Z',
          subscriptionStatus: 'active',
          trialEndsAt: null,
          createdAt: '2024-01-03T17:02:23.290914Z',
        },
      },
    });
  }

  it('will fetch a single currency', async () => {
    givenIAmAuthenticated();
    mockFetch.onGet('/api/locale/currency/JPY').reply(200, yen);

    const world = testRenderHook(() => useCurrency('JPY'), {
      initialRoute: '/settings',
    });
    await waitFor(() => expect(world.result.current.isSuccess).toBeTruthy());
    expect(world.result.current.data).toStrictEqual(yen);
  });

  it('will use the currency list if it is already loaded', async () => {
    givenIAmAuthenticated();
    mockFetch.onGet('/api/locale/currency').reply(200, [euro, yen]);

    const world = testRenderHook(
      ({ code }: { code?: string }) => ({
        list: useInstalledCurrencies(),
        currency: useCurrency(code),
      }),
      {
        initialRoute: '/settings',
        initialProps: { code: undefined } as { code?: string },
      },
    );
    await waitFor(() => expect(world.result.current.list.isSuccess).toBeTruthy());

    world.rerender({ code: 'EUR' });
    expect(world.result.current.currency.data).toStrictEqual(euro);
    // The list is fresh so there should be no need to request the single currency
    expect((mockFetch.history.get ?? []).map(item => item.url)).not.toContain('/api/locale/currency/EUR');
  });

  it('will not fetch without a currency code', () => {
    givenIAmAuthenticated();

    const world = testRenderHook(() => useCurrency(undefined), {
      initialRoute: '/settings',
    });
    expect(world.result.current.data).not.toBeDefined();
    expect(world.result.current.isFetching).toBeFalsy();
    expect(world.result.current.status).toBe('pending');
  });

  it('will not fetch currencies if we are not authenticated', () => {
    mockFetch.onGet('/api/users/me').reply(403, {
      error: 'unauthenticated',
    });

    const world = testRenderHook(() => useCurrency('JPY'), {
      initialRoute: '/login',
    });
    expect(world.result.current.data).not.toBeDefined();
    expect(world.result.current.isFetching).toBeFalsy();
    expect(world.result.current.status).toBe('pending');
  });
});
