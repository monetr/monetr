import React from 'react';
import { waitFor } from '@testing-library/react';
import { rest } from 'msw';

import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import testRenderer from '@monetr/interface/testutils/renderer';
import { server } from '@monetr/interface/testutils/server';

describe('bank sidebar', () => {
  it('will render', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: true,
          allowSignUp: true,
          iconsEnabled: true,
          billingEnabled: false,
        }));
      }),
      rest.get('/api/users/me', (_req, res, ctx) => {
        return res(ctx.json({
          'hasSubscription': true,
          'isActive': true,
          'isSetup': true,
          'user': {
            'account': {
              'accountId': 1,
              'subscriptionActiveUntil': '2023-07-26T00:31:38Z',
              'subscriptionStatus': 'active',
              'timezone': 'America/Chicago',
            },
            'accountId': 1,
            'firstName': 'Elliot',
            'lastName': 'Courant',
            'login': {
              'email': 'email@email.com',
              'emailVerifiedAt': '2022-09-25T00:24:25.976514Z',
              'firstName': 'Elliot',
              'isEmailVerified': true,
              'isPhoneVerified': false,
              'lastName': 'Courant',
              'loginId': 1,
              'passwordResetAt': null,
              'totpEnabledAt': null,
            },
            'loginId': 1,
            'userId': 1,
          },
        }));
      }),
      rest.get('/api/links', (_req, res, ctx) => {
        return res(ctx.json([
          {
            'linkId': 4,
            'linkType': 1,
            'plaidInstitutionId': 'ins_116794',
            'plaidNewAccountsAvailable': false,
            'linkStatus': 2,
            'expirationDate': null,
            'institutionName': 'Mercury',
            'description': null,
            'createdAt': '2022-09-25T02:08:40.758642Z',
            'createdByUserId': 1,
            'updatedAt': '2023-07-02T04:22:52.969206Z',
            'lastManualSync': '2023-05-02T19:56:34.953077Z',
            'lastSuccessfulUpdate': '2023-07-02T04:22:52.96916Z',
          },
        ]));
      }),
      rest.get('/api/institutions/ins_116794', (_req, res, ctx) => {
        return res(ctx.json({
          'institutionId': 'ins_116794',
          'name': 'Mercury',
          'products': [
            'assets',
            'auth',
            'balance',
            'transactions',
            'identity',
          ],
          'countryCodes': [
            'US',
          ],
          'url': 'https://mercury.com/',
          'primaryColor': '#5e6cd4',
          // eslint-disable-next-line max-len
          'logo': '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',
          'status': {
            'auth': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0,
                'success': 1,
              },
              'last_status_change': '2023-06-25T13:10:18Z',
              'status': 'HEALTHY',
            },
            'item_logins': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0.014,
                'success': 0.986,
              },
              'last_status_change': '2023-06-28T20:15:18Z',
              'status': 'DEGRADED',
            },
            'transactions_updates': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0,
                'refresh_interval': 'NORMAL',
                'success': 1,
              },
              'last_status_change': '2023-06-14T17:05:18Z',
              'status': 'HEALTHY',
            },
          },
        }));
      }),
      rest.get('/api/bank_accounts', (_req, res, ctx) => {
        return res(ctx.json([
          {
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
          },
        ]));
      }),
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

    const world = testRenderer(<BankSidebar />, { initialRoute: '/bank/12/transactions' });

    await waitFor(() => expect(world.getByTestId('bank-sidebar')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-settings')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-logout')).toBeVisible());
    await waitFor(() => expect(world.queryByTestId('bank-sidebar-subscription')).not.toBeInTheDocument());

    // Wait for link ID 4 (mercury) to become visible.
    await waitFor(() => expect(world.getByTestId('bank-sidebar-item-4')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-item-4-logo')).toBeVisible());
  });

  it('will show subscription when billing is enabled', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: true,
          allowSignUp: true,
          iconsEnabled: true,
          billingEnabled: true,
        }));
      }),
      rest.get('/api/users/me', (_req, res, ctx) => {
        return res(ctx.json({
          'hasSubscription': true,
          'isActive': true,
          'isSetup': true,
          'user': {
            'account': {
              'accountId': 1,
              'subscriptionActiveUntil': '2023-07-26T00:31:38Z',
              'subscriptionStatus': 'active',
              'timezone': 'America/Chicago',
            },
            'accountId': 1,
            'firstName': 'Elliot',
            'lastName': 'Courant',
            'login': {
              'email': 'email@email.com',
              'emailVerifiedAt': '2022-09-25T00:24:25.976514Z',
              'firstName': 'Elliot',
              'isEmailVerified': true,
              'isPhoneVerified': false,
              'lastName': 'Courant',
              'loginId': 1,
              'passwordResetAt': null,
              'totpEnabledAt': null,
            },
            'loginId': 1,
            'userId': 1,
          },
        }));
      }),
      rest.get('/api/links', (_req, res, ctx) => {
        return res(ctx.json([
          {
            'linkId': 4,
            'linkType': 1,
            'plaidInstitutionId': 'ins_116794',
            'plaidNewAccountsAvailable': false,
            'linkStatus': 2,
            'expirationDate': null,
            'institutionName': 'Mercury',
            'description': null,
            'createdAt': '2022-09-25T02:08:40.758642Z',
            'createdByUserId': 1,
            'updatedAt': '2023-07-02T04:22:52.969206Z',
            'lastManualSync': '2023-05-02T19:56:34.953077Z',
            'lastSuccessfulUpdate': '2023-07-02T04:22:52.96916Z',
          },
        ]));
      }),
      rest.get('/api/institutions/ins_116794', (_req, res, ctx) => {
        return res(ctx.json({
          'institutionId': 'ins_116794',
          'name': 'Mercury',
          'products': [
            'assets',
            'auth',
            'balance',
            'transactions',
            'identity',
          ],
          'countryCodes': [
            'US',
          ],
          'url': 'https://mercury.com/',
          'primaryColor': '#5e6cd4',
          // eslint-disable-next-line max-len
          'logo': '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',
          'status': {
            'auth': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0,
                'success': 1,
              },
              'last_status_change': '2023-06-25T13:10:18Z',
              'status': 'HEALTHY',
            },
            'item_logins': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0.014,
                'success': 0.986,
              },
              'last_status_change': '2023-06-28T20:15:18Z',
              'status': 'DEGRADED',
            },
            'transactions_updates': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0,
                'refresh_interval': 'NORMAL',
                'success': 1,
              },
              'last_status_change': '2023-06-14T17:05:18Z',
              'status': 'HEALTHY',
            },
          },
        }));
      }),
      rest.get('/api/bank_accounts', (_req, res, ctx) => {
        return res(ctx.json([
          {
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
          },
        ]));
      }),
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

    const world = testRenderer(<BankSidebar />, { initialRoute: '/bank/12/transactions' });

    await waitFor(() => expect(world.getByTestId('bank-sidebar')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-settings')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-logout')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-subscription')).toBeVisible());

    // Wait for link ID 4 (mercury) to become visible.
    await waitFor(() => expect(world.getByTestId('bank-sidebar-item-4')).toBeVisible());
  });

  it('will render without an institution icon', async () => {
    server.use(
      rest.get('/api/config', (_req, res, ctx) => {
        return res(ctx.json({
          allowForgotPassword: true,
          allowSignUp: true,
          iconsEnabled: true,
          billingEnabled: false,
        }));
      }),
      rest.get('/api/users/me', (_req, res, ctx) => {
        return res(ctx.json({
          'hasSubscription': true,
          'isActive': true,
          'isSetup': true,
          'user': {
            'account': {
              'accountId': 1,
              'subscriptionActiveUntil': '2023-07-26T00:31:38Z',
              'subscriptionStatus': 'active',
              'timezone': 'America/Chicago',
            },
            'accountId': 1,
            'firstName': 'Elliot',
            'lastName': 'Courant',
            'login': {
              'email': 'email@email.com',
              'emailVerifiedAt': '2022-09-25T00:24:25.976514Z',
              'firstName': 'Elliot',
              'isEmailVerified': true,
              'isPhoneVerified': false,
              'lastName': 'Courant',
              'loginId': 1,
              'passwordResetAt': null,
              'totpEnabledAt': null,
            },
            'loginId': 1,
            'userId': 1,
          },
        }));
      }),
      rest.get('/api/links', (_req, res, ctx) => {
        return res(ctx.json([
          {
            'linkId': 4,
            'linkType': 1,
            'plaidInstitutionId': 'ins_116794',
            'plaidNewAccountsAvailable': false,
            'linkStatus': 2,
            'expirationDate': null,
            'institutionName': 'Mercury',
            'description': null,
            'createdAt': '2022-09-25T02:08:40.758642Z',
            'createdByUserId': 1,
            'updatedAt': '2023-07-02T04:22:52.969206Z',
            'lastManualSync': '2023-05-02T19:56:34.953077Z',
            'lastSuccessfulUpdate': '2023-07-02T04:22:52.96916Z',
          },
        ]));
      }),
      rest.get('/api/institutions/ins_116794', (_req, res, ctx) => {
        return res(ctx.json({
          'institutionId': 'ins_116794',
          'name': 'Mercury',
          'products': [
            'assets',
            'auth',
            'balance',
            'transactions',
            'identity',
          ],
          'countryCodes': [
            'US',
          ],
          'url': 'https://mercury.com/',
          'primaryColor': '#5e6cd4',
          'status': {
            'auth': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0,
                'success': 1,
              },
              'last_status_change': '2023-06-25T13:10:18Z',
              'status': 'HEALTHY',
            },
            'item_logins': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0.014,
                'success': 0.986,
              },
              'last_status_change': '2023-06-28T20:15:18Z',
              'status': 'DEGRADED',
            },
            'transactions_updates': {
              'breakdown': {
                'error_institution': 0,
                'error_plaid': 0,
                'refresh_interval': 'NORMAL',
                'success': 1,
              },
              'last_status_change': '2023-06-14T17:05:18Z',
              'status': 'HEALTHY',
            },
          },
        }));
      }),
      rest.get('/api/bank_accounts', (_req, res, ctx) => {
        return res(ctx.json([
          {
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
          },
        ]));
      }),
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

    const world = testRenderer(<BankSidebar />, { initialRoute: '/bank/12/transactions' });

    await waitFor(() => expect(world.getByTestId('bank-sidebar')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-settings')).toBeVisible());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-logout')).toBeVisible());
    await waitFor(() => expect(world.queryByTestId('bank-sidebar-subscription')).not.toBeInTheDocument());

    // Wait for link ID 4 (mercury) to become visible.
    await waitFor(() => expect(world.getByTestId('bank-sidebar-item-4')).toBeVisible());
    await waitFor(() => expect(world.queryByTestId('bank-sidebar-item-4-logo')).not.toBeInTheDocument());
    await waitFor(() => expect(world.getByTestId('bank-sidebar-item-4-logo-missing')).toBeVisible());
  });
});
