import React, { act, Fragment } from 'react';
import { waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import { showNewBankAccountModal } from '@monetr/interface/modals/NewBankAccountModal';
import testRenderer from '@monetr/interface/testutils/renderer';

const mockUseNavigate = jest.fn((_url: string) => {});
jest.mock('react-router-dom', () => ({
  __esModule: true,
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockUseNavigate,
}));

describe('new bank account modal', () => {
  let mockAxios: MockAdapter;

  beforeEach(() => {
    mockAxios = new MockAdapter(monetrClient);
    mockUseNavigate.mockReset();
  });
  afterEach(() => {
    mockAxios.reset();
  });
  afterAll(() => mockAxios.restore());

  it('will render', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
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
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
      bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
      availableBalance: 47986,
      currentBalance: 47986,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      currency: 'USD',
      lastUpdated: '2024-08-27T08:53:48.555368Z',
      createdAt: '2022-09-25T02:08:40.758642Z',
      updatedAt: '2024-03-19T06:17:32.335106Z',
    });

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions' });
    // Open the dialog
    await act(() => void showNewBankAccountModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-bank-account-modal')).toBeVisible());
    // Close the dialog.
    act(() => world.getByTestId('close-new-bank-account-modal').click());
    // Make sure it goes away.
    await waitFor(() => expect(world.queryByTestId('new-bank-account-modal')).not.toBeInTheDocument());
  });

  it('will attempt to create a new bank account', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
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
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
      bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
      availableBalance: 47986,
      currentBalance: 47986,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      currency: 'USD',
      lastUpdated: '2024-08-27T08:53:48.555368Z',
      createdAt: '2022-09-25T02:08:40.758642Z',
      updatedAt: '2024-03-19T06:17:32.335106Z',
    });

    mockAxios.onPost('/api/bank_accounts').reply(200, {
      bankAccountId: 'bac_created',
      linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
      availableBalance: 10000,
      currentBalance: 10000,
      mask: '',
      name: 'Test Account',
      originalName: 'Test Account',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      currency: 'USD',
    });

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions' });
    // Open the dialog
    await act(() => void showNewBankAccountModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-bank-account-modal')).toBeVisible());

    // Fill out the modal's form and submit it.
    await act(() => userEvent.type(world.getByTestId('bank-account-name'), 'Test Account'));
    await act(() => userEvent.type(world.getByTestId('bank-account-balance'), '100'));
    await act(() => userEvent.click(world.getByTestId('bank-account-submit')));

    // When we submit it we should get redirected to our new bank account.
    await waitFor(() => expect(mockUseNavigate).toBeCalledWith('/bank/bac_created/transactions'));

    // Make sure the modal was also closed.
    await waitFor(() => expect(world.queryByTestId('new-bank-account-modal')).not.toBeInTheDocument());
  });

  it('will create an account with a JPY currency', async () => {
    mockAxios.onGet('/api/users/me').reply(200, {
      activeUntil: '2024-09-26T00:31:38Z',
      hasSubscription: true,
      isActive: true,
      isSetup: true,
      isTrialing: false,
      trialingUntil: null,
      defaultCurrency: 'JPY',
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
          locale: 'ja_JP',
          subscriptionActiveUntil: '2024-09-26T00:31:38Z',
          subscriptionStatus: 'active',
          trialEndsAt: null,
          createdAt: '2024-01-03T17:02:23.290914Z',
        },
      },
    });
    mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
      bankAccountId: 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
      availableBalance: 47986,
      currentBalance: 47986,
      mask: '2982',
      name: 'Mercury Checking',
      originalName: 'Mercury Checking',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      currency: 'JPY',
      lastUpdated: '2024-08-27T08:53:48.555368Z',
      createdAt: '2022-09-25T02:08:40.758642Z',
      updatedAt: '2024-03-19T06:17:32.335106Z',
    });

    mockAxios.onPost('/api/bank_accounts').reply(200, {
      bankAccountId: 'bac_created',
      linkId: 'link_01gds6eqsqacg48p0azb3wcpsq',
      availableBalance: 100,
      currentBalance: 100,
      mask: '',
      name: 'Test Account',
      originalName: 'Test Account',
      accountType: 'depository',
      accountSubType: 'checking',
      status: 'active',
      currency: 'JPY',
    });

    const world = testRenderer(<Fragment />, { initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions' });
    // Open the dialog
    await act(() => void showNewBankAccountModal());
    // Make sure it's visible.
    await waitFor(() => expect(world.getByTestId('new-bank-account-modal')).toBeVisible());

    // Fill out the modal's form and submit it.
    await act(() => userEvent.type(world.getByTestId('bank-account-name'), 'Test Account'));
    await act(() => userEvent.type(world.getByTestId('bank-account-balance'), '100'));
    await act(() => userEvent.click(world.getByTestId('bank-account-submit')));

    // When we submit it we should get redirected to our new bank account.
    await waitFor(() => expect(mockUseNavigate).toBeCalledWith('/bank/bac_created/transactions'));

    // Make sure the modal was also closed.
    await waitFor(() => expect(world.queryByTestId('new-bank-account-modal')).not.toBeInTheDocument());
  });
});
