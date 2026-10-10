import { Route } from 'wouter';

import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import LinkDetails from '@monetr/interface/pages/link/details';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('link details page', () => {
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

  it('will render a plaid link with its accounts in the saved order', async () => {
    mockFetch.onGet('/api/users/me').reply(200, {
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      defaultCurrency: 'USD',
      user: {
        account: {
          accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
          subscriptionActiveUntil: '2026-11-26T00:31:38Z',
          subscriptionStatus: 'active',
          timezone: 'America/Chicago',
          locale: 'en_US',
        },
        accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
        login: {
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          isEmailVerified: true,
          loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
          totpEnabledAt: null,
        },
        loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        role: 'owner',
        linkOrder: null,
      },
    });
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'plaid',
      plaidLink: {
        products: ['transactions'],
        status: 'setup',
        errorCode: null,
        expirationDate: null,
        newAccountsAvailable: false,
        institutionId: 'ins_116794',
        institutionName: 'Mercury',
        lastManualSync: '2026-10-08T19:56:34Z',
        lastSuccessfulUpdate: '2026-10-10T04:22:52Z',
        lastAttemptedUpdate: '2026-10-10T04:22:52Z',
        updatedAt: '2026-10-10T04:22:52Z',
        createdAt: '2026-01-12T02:08:40Z',
        createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      },
      institutionName: 'Mercury',
      description: null,
      bankAccountOrder: ['bac_01hy4rcmadc01d2kzv7vynbyyy', 'bac_01hy4rcmadc01d2kzv7vynbxxx'],
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T04:22:52Z',
      deletedAt: null,
    });
    mockFetch.onGet('/api/bank_accounts?link_id=link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, [
      {
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
        lastUpdated: '2026-10-10T04:22:52Z',
        createdAt: '2026-01-12T02:08:40Z',
        deletedAt: null,
      },
      {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbyyy',
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
        availableBalance: 812000,
        currentBalance: 812000,
        mask: '4410',
        name: 'Mercury Savings',
        originalName: 'Mercury Savings',
        accountType: 'depository',
        accountSubType: 'savings',
        status: 'active',
        currency: 'USD',
        lastUpdated: '2026-10-10T04:22:52Z',
        createdAt: '2026-01-12T02:08:40Z',
        deletedAt: null,
      },
      {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbzzz',
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
        availableBalance: 0,
        currentBalance: 0,
        mask: '1187',
        name: 'Old Joint Checking',
        originalName: 'Old Joint Checking',
        accountType: 'depository',
        accountSubType: 'checking',
        status: 'active',
        currency: 'USD',
        lastUpdated: '2026-02-01T04:22:52Z',
        createdAt: '2026-01-12T02:08:40Z',
        deletedAt: '2026-02-01T04:22:52Z',
      },
    ]);

    const world = testRenderer(<Route component={LinkDetails} path='/link/:linkId/details' />, {
      initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
    });

    await waitFor(() => expect(world.getByTestId('link-status')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('link-status-sync')).toBeVisible());
    await waitFor(() => expect(world.queryByTestId('link-alert-error')).not.toBeInTheDocument());

    // Savings was dragged above checking, and the archived account stays out of the sortable list.
    await waitFor(() =>
      expect(
        Array.from(world.getByTestId('link-accounts').children).map(item => item.getAttribute('data-sortable-id')),
      ).toEqual(['bac_01hy4rcmadc01d2kzv7vynbyyy', 'bac_01hy4rcmadc01d2kzv7vynbxxx']),
    );
    await waitFor(() => expect(world.getByTestId('link-accounts-archived')).toHaveTextContent('Old Joint Checking'));
  });

  it('will render a manual link without a sync status', async () => {
    mockFetch.onGet('/api/users/me').reply(200, {
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      defaultCurrency: 'USD',
      user: {
        account: {
          accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
          subscriptionActiveUntil: '2026-11-26T00:31:38Z',
          subscriptionStatus: 'active',
          timezone: 'America/Chicago',
          locale: 'en_US',
        },
        accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
        login: {
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          isEmailVerified: true,
          loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
          totpEnabledAt: null,
        },
        loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        role: 'owner',
        linkOrder: null,
      },
    });
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Cash Budget',
      description: null,
      bankAccountOrder: null,
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T04:22:52Z',
      deletedAt: null,
    });
    mockFetch.onGet('/api/bank_accounts?link_id=link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, [
      {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
        availableBalance: 12000,
        currentBalance: 12000,
        mask: null,
        name: 'Envelope Cash',
        originalName: 'Envelope Cash',
        accountType: 'depository',
        accountSubType: 'checking',
        status: 'active',
        currency: 'USD',
        lastUpdated: '2026-10-10T04:22:52Z',
        createdAt: '2026-01-12T02:08:40Z',
        deletedAt: null,
      },
    ]);

    const world = testRenderer(<Route component={LinkDetails} path='/link/:linkId/details' />, {
      initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
    });

    await waitFor(() => expect(world.getByTestId('link-details-name')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('link-accounts')).toHaveTextContent('Envelope Cash'));
    await waitFor(() => expect(world.queryByTestId('link-status')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('link-status-sync')).not.toBeInTheDocument());
    await waitFor(() => expect(world.queryByTestId('link-accounts-archived')).not.toBeInTheDocument());
  });

  it('will show an alert when the plaid link needs to sign in again', async () => {
    mockFetch.onGet('/api/users/me').reply(200, {
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      defaultCurrency: 'USD',
      user: {
        account: {
          accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
          subscriptionActiveUntil: '2026-11-26T00:31:38Z',
          subscriptionStatus: 'active',
          timezone: 'America/Chicago',
          locale: 'en_US',
        },
        accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
        login: {
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          isEmailVerified: true,
          loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
          totpEnabledAt: null,
        },
        loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        role: 'owner',
        linkOrder: null,
      },
    });
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'plaid',
      plaidLink: {
        products: ['transactions'],
        status: 'error',
        errorCode: 'ITEM_LOGIN_REQUIRED',
        expirationDate: null,
        newAccountsAvailable: false,
        institutionId: 'ins_116794',
        institutionName: 'Mercury',
        lastManualSync: '2026-10-05T19:56:34Z',
        lastSuccessfulUpdate: '2026-10-06T23:20:00Z',
        lastAttemptedUpdate: '2026-10-10T04:22:52Z',
        updatedAt: '2026-10-10T04:22:52Z',
        createdAt: '2026-01-12T02:08:40Z',
        createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      },
      institutionName: 'Mercury',
      description: null,
      bankAccountOrder: null,
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T04:22:52Z',
      deletedAt: null,
    });
    mockFetch.onGet('/api/bank_accounts?link_id=link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, [
      {
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
        lastUpdated: '2026-10-06T23:20:00Z',
        createdAt: '2026-01-12T02:08:40Z',
        deletedAt: null,
      },
    ]);

    const world = testRenderer(<Route component={LinkDetails} path='/link/:linkId/details' />, {
      initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
    });

    await waitFor(() => expect(world.getByTestId('link-alert-error')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('link-status')).toHaveTextContent('Needs Attention'));
    await waitFor(() => expect(world.queryByTestId('link-alert-expiring')).not.toBeInTheDocument());
  });

  it('will save the name and description', async () => {
    mockFetch.onGet('/api/users/me').reply(200, {
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      defaultCurrency: 'USD',
      user: {
        account: {
          accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
          subscriptionActiveUntil: '2026-11-26T00:31:38Z',
          subscriptionStatus: 'active',
          timezone: 'America/Chicago',
          locale: 'en_US',
        },
        accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
        login: {
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          isEmailVerified: true,
          loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
          totpEnabledAt: null,
        },
        loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        role: 'owner',
        linkOrder: null,
      },
    });
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Cash Budget',
      description: null,
      bankAccountOrder: null,
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T04:22:52Z',
      deletedAt: null,
    });
    mockFetch.onGet('/api/bank_accounts?link_id=link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, []);
    mockFetch.onPatch('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Envelope Budget',
      description: 'The cash we keep at home',
      bankAccountOrder: null,
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T05:00:00Z',
      deletedAt: null,
    });

    const world = testRenderer(<Route component={LinkDetails} path='/link/:linkId/details' />, {
      initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
    });

    await waitFor(() => expect(world.getByTestId('link-details-name')).toBeVisible());

    // Rename the link and give it a description
    const user = userEvent.setup();
    await user.clear(world.getByTestId('link-details-name'));
    await user.type(world.getByTestId('link-details-name'), 'Envelope Budget');
    await user.type(world.getByTestId('link-details-description'), 'The cash we keep at home');
    await user.click(world.getByTestId('link-details-save'));

    await waitFor(() => expect(mockFetch.history.patch).toHaveLength(1));
    expect(mockFetch.history.patch?.[0]?.data).toMatchObject({
      institutionName: 'Envelope Budget',
      description: 'The cash we keep at home',
    });
  });

  it('will clear the description when it is emptied', async () => {
    mockFetch.onGet('/api/users/me').reply(200, {
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      defaultCurrency: 'USD',
      user: {
        account: {
          accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
          subscriptionActiveUntil: '2026-11-26T00:31:38Z',
          subscriptionStatus: 'active',
          timezone: 'America/Chicago',
          locale: 'en_US',
        },
        accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
        login: {
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          isEmailVerified: true,
          loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
          totpEnabledAt: null,
        },
        loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        role: 'owner',
        linkOrder: null,
      },
    });
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Cash Budget',
      description: 'The cash we keep at home',
      bankAccountOrder: null,
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T04:22:52Z',
      deletedAt: null,
    });
    mockFetch.onGet('/api/bank_accounts?link_id=link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, []);
    mockFetch.onPatch('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'manual',
      institutionName: 'Cash Budget',
      description: null,
      bankAccountOrder: null,
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T05:00:00Z',
      deletedAt: null,
    });

    const world = testRenderer(<Route component={LinkDetails} path='/link/:linkId/details' />, {
      initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
    });

    await waitFor(() => expect(world.getByTestId('link-details-description')).toHaveValue('The cash we keep at home'));

    // Empty out the description and save
    const user = userEvent.setup();
    await user.clear(world.getByTestId('link-details-description'));
    await user.click(world.getByTestId('link-details-save'));

    // The API only clears the description on null, an empty string would get stored.
    await waitFor(() => expect(mockFetch.history.patch).toHaveLength(1));
    expect(mockFetch.history.patch?.[0]?.data).toMatchObject({
      institutionName: 'Cash Budget',
      description: null,
    });
  });

  it('will trigger a manual sync', async () => {
    mockFetch.onGet('/api/users/me').reply(200, {
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      defaultCurrency: 'USD',
      user: {
        account: {
          accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
          subscriptionActiveUntil: '2026-11-26T00:31:38Z',
          subscriptionStatus: 'active',
          timezone: 'America/Chicago',
          locale: 'en_US',
        },
        accountId: 'acct_01hy4r9b8gnb1a1d2r0z6t2k8x',
        login: {
          email: 'example@example.com',
          firstName: 'Elliot',
          lastName: 'Courant',
          isEmailVerified: true,
          loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
          totpEnabledAt: null,
        },
        loginId: 'lgn_01hy4r9b8gnb1a1d2r0z6t2k8x',
        userId: 'user_01hym36e8ewaq0hxssb1m3k4ha',
        role: 'owner',
        linkOrder: null,
      },
    });
    mockFetch.onGet('/api/links/link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, {
      linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      linkType: 'plaid',
      plaidLink: {
        products: ['transactions'],
        status: 'setup',
        errorCode: null,
        expirationDate: null,
        newAccountsAvailable: false,
        institutionId: 'ins_116794',
        institutionName: 'Mercury',
        lastManualSync: '2026-10-08T19:56:34Z',
        lastSuccessfulUpdate: '2026-10-10T04:22:52Z',
        lastAttemptedUpdate: '2026-10-10T04:22:52Z',
        updatedAt: '2026-10-10T04:22:52Z',
        createdAt: '2026-01-12T02:08:40Z',
        createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      },
      institutionName: 'Mercury',
      description: null,
      bankAccountOrder: null,
      createdAt: '2026-01-12T02:08:40Z',
      createdBy: 'user_01hym36e8ewaq0hxssb1m3k4ha',
      updatedAt: '2026-10-10T04:22:52Z',
      deletedAt: null,
    });
    mockFetch.onGet('/api/bank_accounts?link_id=link_01hy4rbb1gjdek7h2xmgy5pnwk').reply(200, []);
    mockFetch.onPost('/api/plaid/link/sync').reply(200);

    const world = testRenderer(<Route component={LinkDetails} path='/link/:linkId/details' />, {
      initialRoute: '/link/link_01hy4rbb1gjdek7h2xmgy5pnwk/details',
    });

    await waitFor(() => expect(world.getByTestId('link-status-sync')).toBeVisible());

    // Kick off the sync
    const user = userEvent.setup();
    await user.click(world.getByTestId('link-status-sync'));

    await waitFor(() => expect(mockFetch.history.post).toHaveLength(1));
    expect(mockFetch.history.post?.[0]).toMatchObject({
      url: '/api/plaid/link/sync',
      data: {
        linkId: 'link_01hy4rbb1gjdek7h2xmgy5pnwk',
      },
    });
  });
});
