/* eslint-disable max-len */
import React from 'react';
import { Meta, StoryObj } from '@storybook/react';

import MonetrWrapper, { BankView, ExpensesView, TransactionsView } from './new';

import { rest } from 'msw';

const meta: Meta<typeof MonetrWrapper> = {
  title: 'New UI',
  component: MonetrWrapper,
  parameters: {
    msw: {
      handlers: [
        rest.get('/api/config', (_req, res, ctx) => {
          return res(ctx.json({
            allowForgotPassword: true,
            allowSignUp: true,
            iconsEnabled: true,
          }));
        }),
        rest.get('/api/institutions/ins_15', (_req, res, ctx) => {
          return res(ctx.json({
            'institutionId': 'ins_15',
            'name': 'Navy Federal Credit Union',
            'products': [
              'assets',
              'auth',
              'balance',
              'transactions',
              'identity',
              'investments',
              'liabilities',
            ],
            'countryCodes': [
              'US',
            ],
            'url': 'https://www.navyfederal.org/',
            'primaryColor': '#0056a1',
            // eslint-disable-next-line max-len
            'logo': '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',
            'status': {
              'auth': {
                'breakdown': {
                  'error_institution': 0.002,
                  'error_plaid': 0.016,
                  'success': 0.982,
                },
                'last_status_change': '2023-06-28T21:45:18Z',
                'status': 'DEGRADED',
              },
              'health_incidents': [
                {
                  'end_date': '2023-06-08T18:29:00Z',
                  'incident_updates': [
                    {
                      'description': 'On May 16, 2023, Navy Federal Credit Union (NFCU) made technical changes to their online banking platform resulting in all NFCU customers needing to re-authenticate via Plaid in order to maintain ongoing access to connected apps.  Items will move into an ITEM_LOGIN_REQUIRED error state and trigger an ITEM ERROR webhook. Users should relink their accounts via Link Update Mode.\n',
                      'status': 'SCHEDULED',
                      'updated_date': '2023-06-02T19:35:51Z',
                    },
                  ],
                  'start_date': '2023-05-16T18:29:00Z',
                  'title': 'All NFCU customers must re-authenticate their connections',
                },
                {
                  'end_date': '2023-06-08T20:28:00Z',
                  'incident_updates': [
                    {
                      'description': 'Some users may encounter an erroneous USER_SETUP_REQUIRED error when trying to connect their NFCU accounts.  Our engineering team is aware of this issue and is working on a fix.',
                      'status': 'INVESTIGATING',
                      'updated_date': '2023-06-01T21:01:51Z',
                    },
                  ],
                  'start_date': '2023-05-24T20:28:00Z',
                  'title': 'Erroneous USER_SETUP_REQUIRED error',
                },
              ],
              'identity': {
                'breakdown': {
                  'error_institution': 0.012,
                  'error_plaid': 0.002,
                  'success': 0.985,
                },
                'last_status_change': '2023-06-28T18:50:18Z',
                'status': 'DEGRADED',
              },
              'investments_updates': {
                'breakdown': {
                  'error_institution': 0.001,
                  'error_plaid': 0.134,
                  'refresh_interval': 'STOPPED',
                  'success': 0.864,
                },
                'last_status_change': '2023-06-28T22:20:18Z',
                'status': 'DOWN',
              },
              'item_logins': {
                'breakdown': {
                  'error_institution': 0.012,
                  'error_plaid': 0.022,
                  'success': 0.966,
                },
                'last_status_change': '2023-06-28T01:35:18Z',
                'status': 'DEGRADED',
              },
              'liabilities_updates': {
                'breakdown': {
                  'error_institution': 0.002,
                  'error_plaid': 0.004,
                  'refresh_interval': 'NORMAL',
                  'success': 0.99,
                },
                'last_status_change': '2023-06-28T21:40:18Z',
                'status': 'HEALTHY',
              },
              'transactions_updates': {
                'breakdown': {
                  'error_institution': 0.002,
                  'error_plaid': 0.001,
                  'refresh_interval': 'NORMAL',
                  'success': 0.995,
                },
                'last_status_change': '2023-06-28T21:40:18Z',
                'status': 'HEALTHY',
              },
            },
          }));
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
        rest.get('/api/institutions/ins_127990', (_req, res, ctx) => {
          return res(ctx.json({
            'countryCodes': [
              'US',
            ],
            'institutionId': 'ins_127990',
            'logo': 'iVBORw0KGgoAAAANSUhEUgAAAJgAAACYCAMAAAAvHNATAAAAQlBMVEVHcEz////////////////////////////////////////////////////////////bGzLujpndJDryp7D98PHmW2vEUBy/AAAAD3RSTlMAqqCBEu3QIeC/ljVp+kt/eKdWAAADSUlEQVR42u2caVbrMAyFG8dz7YhS2P9WefA4hbRJPUf3B3cF3/GgyIp0T6dGyfMyi8k5bVWMymrnJjEvZ3lilAneWdqRdT4YFiixy/SLThwLt/gMqBucXw5aq7mA6pttHr5uMmiqkg4jr4PxkaoV/ahlM6IB6wtNjEA7T9RB07n3anXB+ly1qeuqBUXdpEK/XdTUVbrPfkpB3SU6xI5F0QCp5q/BTIM0t11GR8PkDNo2tm/nTINVuZ2ehsvXcE10gKby6OXoELnCiCY1HSRdRCYtHSYrEdercM0cHSoHdR8r7qanw+Uh4n3lN2AhFiW/m0axcEVlsC5k9tWciU0z4AFLHjOpGLmi2v8CCGKV2H0/ErP23puaG0zv1AGIXQEotK4vgEHIKTZrQYAnf+/8TxhgD0tmMLgeT5kAAbuPsiYS5pJ5gtEqzZYRByxKrKC/Hf41EpjOixWX1xG6PCMzWQn15WWELnlJtsUCs1mZPgPYLfv3aGA+YydZwGzO95sD7PteBjywkJFYsICJ9BHjAbMZKSIL2NchC4hgIZ2K8YD5dEWMB8wlzz4T2L/TLwkRjGTyocsEdk4WEZnAlmTVlQlsTr50mcBEsmbBBDYlC/tMYA4XTGOC6VTg5wKzJ4UJpk4REyzigsFuJezhhw0Xf5G/FAz2Iw6b9sAmirCpNexjBPb5BvvgxS0RwBZVYMtQsIU72FInbnEYtpwO+wMC9pfN80h2HQJ2zfrJ9TzBeLu+d6Z6v76lUouc32//2V4Po/r1IzWnl7MPWwbV6tdzXmNPK1se1epnfXZ7Qz1bNtW6IbCgIaSGrYDqriGkrIWmjK2M6q6FprjpKJetlOq+6aimTSvNVkH12NlW1dj2jK2Kih7bh2tbAbfZaqnWsaK1efKerYFqa66lqd30h62JijbbrRv7TT/ZWqm2B4EQWpq3J0dQm8Bx2+ZxBw1gRzN4h1noyTAL7PgP7sAU7ogZW5hNDuXBjjHiDn7ijsriDhfjjmPjDrDDjvzjmiTg2krgGnEAW5fgmr3g2uPgGgrhWjABm1bh2nzhGqP1tpKL/azkeprv0dTbshDUrhDY4BHYEhPYRBTYdhXZqBbZ2vcGB2iG/HMdBtlHfwCEijqNZ6fsZgAAAABJRU5ErkJggg==',
            'name': 'U.S. Bank',
            'primaryColor': '#102670',
            'products': [
              'assets',
              'auth',
              'balance',
              'transactions',
              'credit_details',
              'income',
              'identity',
              'investments',
              'liabilities',
            ],
            'status': {
              'auth': {
                'breakdown': {
                  'error_institution': 0.004,
                  'error_plaid': 0.002,
                  'success': 0.994,
                },
                'last_status_change': '2023-06-28T02:15:18Z',
                'status': 'HEALTHY',
              },
              'identity': {
                'breakdown': {
                  'error_institution': 0,
                  'error_plaid': 0,
                  'success': 1,
                },
                'last_status_change': '2023-05-07T11:45:18Z',
                'status': 'HEALTHY',
              },
              'investments_updates': {
                'breakdown': {
                  'error_institution': 0.006,
                  'error_plaid': 0.327,
                  'refresh_interval': 'NORMAL',
                  'success': 0.664,
                },
                'last_status_change': '2023-06-29T00:30:18Z',
                'status': 'DEGRADED',
              },
              'item_logins': {
                'breakdown': {
                  'error_institution': 0.03,
                  'error_plaid': 0.008,
                  'success': 0.963,
                },
                'last_status_change': '2023-05-05T21:00:18Z',
                'status': 'DEGRADED',
              },
              'liabilities_updates': {
                'breakdown': {
                  'error_institution': 0.009,
                  'error_plaid': 0.009,
                  'refresh_interval': 'NORMAL',
                  'success': 0.981,
                },
                'last_status_change': '2023-06-29T00:30:18Z',
                'status': 'DEGRADED',
              },
              'transactions_updates': {
                'breakdown': {
                  'error_institution': 0.001,
                  'error_plaid': 0.002,
                  'refresh_interval': 'NORMAL',
                  'success': 0.995,
                },
                'last_status_change': '2023-06-29T00:25:18Z',
                'status': 'HEALTHY',
              },
            },
            'url': 'https://www.usbank.com/',
          }));
        }),
        rest.get('/api/institutions/ins_3', (_req, res, ctx) => {
          return res(ctx.json({
            'countryCodes': [
              'US',
            ],
            'institutionId': 'ins_3',
            'name': 'Chase',
            'primaryColor': '#095aa6',
            'products': [
              'assets',
              'auth',
              'balance',
              'transactions',
              'credit_details',
              'income',
              'identity',
              'investments',
              'liabilities',
            ],
            'status': {
              'health_incidents': [
                {
                  'end_date': '2021-11-07T11:03:00Z',
                  'incident_updates': [
                    {
                      'description': 'Chase will be conducting maintenance on all channels during the following time:  November 6th @ 23:30 PM ending November 7th @ 4:00 AM EST During this time all attempts to access Chase will be blocked.   Tokenized Channels will receive a “Server Unavailable” error.  Once the window ends, all activities will  return to normal',
                      'status': 'SCHEDULED',
                      'updated_date': '2021-11-05T22:24:15Z',
                    },
                  ],
                  'start_date': '2021-11-05T22:02:00Z',
                  'title': 'Open Banking Scheduled Downtime',
                },
                {
                  'incident_updates': [
                    {
                      'description': 'You\'re viewing status data for ins_3, our legacy Chase integration. This data is relevant for existing ins_3 items. To view status data for our new Chase OAuth integration, please see status information for ins_56. If you are enabled, all new Chase item adds will be routed through ins_56. If you have questions please contact Plaid Support.',
                      'status': 'SCHEDULED',
                      'updated_date': '2021-12-02T04:25:07Z',
                    },
                  ],
                  'start_date': '2021-12-02T04:23:00Z',
                  'title': 'Institution Deprecation',
                },
              ],
              'item_logins': {
                'breakdown': {
                  'error_institution': 0,
                  'error_plaid': 0,
                  'success': 1,
                },
                'last_status_change': '2023-05-14T12:50:18Z',
                'status': 'HEALTHY',
              },
            },
            'url': 'https://www.chase.com',
          }));
        }),
        rest.post('/api/icons/search', async (req, res, ctx) => {
          const body = await req.json();
          switch (body['name']) {
            case 'Discord':
              return res(ctx.json({
                'colors': [
                  '5865F2',
                ],
                'library': 'simple-icons',
                'slug': 'discord',
                'svg': 'PHN2ZyByb2xlPSJpbWciIHZpZXdCb3g9IjAgMCAyNCAyNCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48dGl0bGU+RGlzY29yZDwvdGl0bGU+PHBhdGggZD0iTTIwLjMxNyA0LjM2OThhMTkuNzkxMyAxOS43OTEzIDAgMDAtNC44ODUxLTEuNTE1Mi4wNzQxLjA3NDEgMCAwMC0uMDc4NS4wMzcxYy0uMjExLjM3NTMtLjQ0NDcuODY0OC0uNjA4MyAxLjI0OTUtMS44NDQ3LS4yNzYyLTMuNjgtLjI3NjItNS40ODY4IDAtLjE2MzYtLjM5MzMtLjQwNTgtLjg3NDItLjYxNzctMS4yNDk1YS4wNzcuMDc3IDAgMDAtLjA3ODUtLjAzNyAxOS43MzYzIDE5LjczNjMgMCAwMC00Ljg4NTIgMS41MTUuMDY5OS4wNjk5IDAgMDAtLjAzMjEuMDI3N0MuNTMzNCA5LjA0NTgtLjMxOSAxMy41Nzk5LjA5OTIgMTguMDU3OGEuMDgyNC4wODI0IDAgMDAuMDMxMi4wNTYxYzIuMDUyOCAxLjUwNzYgNC4wNDEzIDIuNDIyOCA1Ljk5MjkgMy4wMjk0YS4wNzc3LjA3NzcgMCAwMC4wODQyLS4wMjc2Yy40NjE2LS42MzA0Ljg3MzEtMS4yOTUyIDEuMjI2LTEuOTk0MmEuMDc2LjA3NiAwIDAwLS4wNDE2LS4xMDU3Yy0uNjUyOC0uMjQ3Ni0xLjI3NDMtLjU0OTUtMS44NzIyLS44OTIzYS4wNzcuMDc3IDAgMDEtLjAwNzYtLjEyNzdjLjEyNTgtLjA5NDMuMjUxNy0uMTkyMy4zNzE4LS4yOTE0YS4wNzQzLjA3NDMgMCAwMS4wNzc2LS4wMTA1YzMuOTI3OCAxLjc5MzMgOC4xOCAxLjc5MzMgMTIuMDYxNCAwYS4wNzM5LjA3MzkgMCAwMS4wNzg1LjAwOTVjLjEyMDIuMDk5LjI0Ni4xOTgxLjM3MjguMjkyNGEuMDc3LjA3NyAwIDAxLS4wMDY2LjEyNzYgMTIuMjk4NiAxMi4yOTg2IDAgMDEtMS44NzMuODkxNC4wNzY2LjA3NjYgMCAwMC0uMDQwNy4xMDY3Yy4zNjA0LjY5OC43NzE5IDEuMzYyOCAxLjIyNSAxLjk5MzJhLjA3Ni4wNzYgMCAwMC4wODQyLjAyODZjMS45NjEtLjYwNjcgMy45NDk1LTEuNTIxOSA2LjAwMjMtMy4wMjk0YS4wNzcuMDc3IDAgMDAuMDMxMy0uMDU1MmMuNTAwNC01LjE3Ny0uODM4Mi05LjY3MzktMy41NDg1LTEzLjY2MDRhLjA2MS4wNjEgMCAwMC0uMDMxMi0uMDI4NnpNOC4wMiAxNS4zMzEyYy0xLjE4MjUgMC0yLjE1NjktMS4wODU3LTIuMTU2OS0yLjQxOSAwLTEuMzMzMi45NTU1LTIuNDE4OSAyLjE1Ny0yLjQxODkgMS4yMTA4IDAgMi4xNzU3IDEuMDk1MiAyLjE1NjggMi40MTkgMCAxLjMzMzItLjk1NTUgMi40MTg5LTIuMTU2OSAyLjQxODl6bTcuOTc0OCAwYy0xLjE4MjUgMC0yLjE1NjktMS4wODU3LTIuMTU2OS0yLjQxOSAwLTEuMzMzMi45NTU0LTIuNDE4OSAyLjE1NjktMi40MTg5IDEuMjEwOCAwIDIuMTc1NyAxLjA5NTIgMi4xNTY4IDIuNDE5IDAgMS4zMzMyLS45NDYgMi40MTg5LTIuMTU2OCAyLjQxODlaIi8+PC9zdmc+',
                'title': 'discord',
              }));
            case 'Target':
              return res(ctx.json({
                'colors': [
                  'CC0000',
                ],
                'library': 'simple-icons',
                'slug': 'target',
                'svg': 'PHN2ZyByb2xlPSJpbWciIHZpZXdCb3g9IjAgMCAyNCAyNCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48dGl0bGU+VGFyZ2V0PC90aXRsZT48cGF0aCBkPSJNMTIuMDAwNSAwQzE4LjYyNyAwIDI0IDUuMzczIDI0IDEyLjAwMDUgMjQgMTguNjI3IDE4LjYyNyAyNCAxMS45OTk1IDI0IDUuMzczIDI0IDAgMTguNjI3IDAgMTEuOTk5NSAwIDUuMzczIDUuMzczIDAgMTIuMDAwNSAwem0wIDE5LjgyNmE3LjgyNjUgNy44MjY1IDAgMTAtLjAwMS0xNS42NTJDNy43MTMzIDQuMjI0NiA0LjI2NTMgNy43MTM2IDQuMjY1MyAxMmMwIDQuMjg2NCAzLjQ0OCA3Ljc3NTQgNy43MzQyIDcuODI2aC4wMDF6bTAtMy45ODUzYTMuODQwMiAzLjg0MDIgMCAxMTAtNy42ODAzYzIuMTIwNC4wMDA2IDMuODM5IDEuNzE5NyAzLjgzOSAzLjg0MDFzLTEuNzE4NiAzLjgzOTYtMy44MzkgMy44NDAyeiIvPjwvc3ZnPg==',
                'title': 'target',
              }));
            case 'GitHub':
              return res(ctx.json({
                'colors': [
                  '181717',
                ],
                'library': 'simple-icons',
                'slug': 'github',
                'svg': 'PHN2ZyByb2xlPSJpbWciIHZpZXdCb3g9IjAgMCAyNCAyNCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48dGl0bGU+R2l0SHViPC90aXRsZT48cGF0aCBkPSJNMTIgLjI5N2MtNi42MyAwLTEyIDUuMzczLTEyIDEyIDAgNS4zMDMgMy40MzggOS44IDguMjA1IDExLjM4NS42LjExMy44Mi0uMjU4LjgyLS41NzcgMC0uMjg1LS4wMS0xLjA0LS4wMTUtMi4wNC0zLjMzOC43MjQtNC4wNDItMS42MS00LjA0Mi0xLjYxQzQuNDIyIDE4LjA3IDMuNjMzIDE3LjcgMy42MzMgMTcuN2MtMS4wODctLjc0NC4wODQtLjcyOS4wODQtLjcyOSAxLjIwNS4wODQgMS44MzggMS4yMzYgMS44MzggMS4yMzYgMS4wNyAxLjgzNSAyLjgwOSAxLjMwNSAzLjQ5NS45OTguMTA4LS43NzYuNDE3LTEuMzA1Ljc2LTEuNjA1LTIuNjY1LS4zLTUuNDY2LTEuMzMyLTUuNDY2LTUuOTMgMC0xLjMxLjQ2NS0yLjM4IDEuMjM1LTMuMjItLjEzNS0uMzAzLS41NC0xLjUyMy4xMDUtMy4xNzYgMCAwIDEuMDA1LS4zMjIgMy4zIDEuMjMuOTYtLjI2NyAxLjk4LS4zOTkgMy0uNDA1IDEuMDIuMDA2IDIuMDQuMTM4IDMgLjQwNSAyLjI4LTEuNTUyIDMuMjg1LTEuMjMgMy4yODUtMS4yMy42NDUgMS42NTMuMjQgMi44NzMuMTIgMy4xNzYuNzY1Ljg0IDEuMjMgMS45MSAxLjIzIDMuMjIgMCA0LjYxLTIuODA1IDUuNjI1LTUuNDc1IDUuOTIuNDIuMzYuODEgMS4wOTYuODEgMi4yMiAwIDEuNjA2LS4wMTUgMi44OTYtLjAxNSAzLjI4NiAwIC4zMTUuMjEuNjkuODI1LjU3QzIwLjU2NSAyMi4wOTIgMjQgMTcuNTkyIDI0IDEyLjI5N2MwLTYuNjI3LTUuMzczLTEyLTEyLTEyIi8+PC9zdmc+',
                'title': 'github',
              }));
            default:
              return res(ctx.status(204));
          }
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
        rest.get('/api/bank_accounts/12/balances', (_req, res, ctx) => {
          return res(ctx.json({
            'bankAccountId': 12,
            'current': 48635,
            'available': 48635,
            'free': -1345,
            'expenses': 49311,
            'goals': 669,
          }));
        }),
        rest.get('/api/bank_accounts/12/spending', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'spendingId': 59,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Domains',
              'description': 'Every year on the 1st of April',
              'targetAmount': 14000,
              'currentAmount': 1998,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=YEARLY;INTERVAL=1;BYMONTH=4;BYMONTHDAY=1',
              'lastRecurrence': null,
              'nextRecurrence': '2024-04-01T05:00:00Z',
              'nextContributionAmount': 666,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2023-05-16T14:47:58.09301Z',
              'dateStarted': '2024-04-01T05:00:00Z',
            },
            {
              'spendingId': 58,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Plaid',
              'description': 'Every month on the 10th',
              'targetAmount': 500,
              'currentAmount': 500,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10',
              'lastRecurrence': '2023-07-10T05:00:00Z',
              'nextRecurrence': '2023-07-10T05:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2023-05-14T20:01:47.09268Z',
              'dateStarted': '2023-06-10T05:00:00Z',
            },
            {
              'spendingId': 63,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'GitLab',
              'description': 'Every year on the 16th of June',
              'targetAmount': 34800,
              'currentAmount': 1450,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=YEARLY;INTERVAL=1;BYMONTH=6;BYMONTHDAY=16',
              'lastRecurrence': '2024-06-16T05:00:00Z',
              'nextRecurrence': '2024-06-16T05:00:00Z',
              'nextContributionAmount': 1450,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2023-06-05T16:07:02.780623Z',
              'dateStarted': '2023-06-06T05:00:00Z',
            },
            {
              'spendingId': 192,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Cloud Staging',
              'description': 'Every month on the 1st',
              'targetAmount': 10000,
              'currentAmount': 10000,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-08-01T05:00:00Z',
              'nextRecurrence': '2023-08-01T05:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-07T15:09:32Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            },
            {
              'spendingId': 189,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Google Voice',
              'description': 'Every month on the 1st',
              'targetAmount': 1366,
              'currentAmount': 1366,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-08-01T05:00:00Z',
              'nextRecurrence': '2023-08-01T05:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-02T14:11:24Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            },
            {
              'spendingId': 201,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 1,
              'name': 'Rainy Day',
              'targetAmount': 100,
              'currentAmount': 669,
              'usedAmount': 0,
              'recurrenceRule': null,
              'lastRecurrence': '2022-12-31T06:00:00Z',
              'nextRecurrence': '2023-12-31T06:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-29T16:32:58Z',
              'dateStarted': '2022-12-31T06:00:00Z',
            },
            {
              'spendingId': 208,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Google Domains ($12)',
              'description': 'Every year on the 29th of January',
              'targetAmount': 1200,
              'currentAmount': 547,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=YEARLY;INTERVAL=1;BYMONTH=1;BYMONTHDAY=29',
              'lastRecurrence': null,
              'nextRecurrence': '2024-01-29T06:00:00Z',
              'nextContributionAmount': 50,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2023-01-30T22:16:20Z',
              'dateStarted': '2024-01-29T06:00:00Z',
            },
            {
              'spendingId': 138,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'GitHub',
              'description': 'Every month on the 19th',
              'targetAmount': 2600,
              'currentAmount': 1300,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=19',
              'lastRecurrence': '2023-07-19T05:00:00Z',
              'nextRecurrence': '2023-07-19T05:00:00Z',
              'nextContributionAmount': 1300,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:43:04Z',
              'dateStarted': '2023-03-19T05:00:00Z',
            },
            {
              'spendingId': 136,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'BuildKite',
              'description': 'Every month on the 27th',
              'targetAmount': 1500,
              'currentAmount': 750,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=27',
              'lastRecurrence': '2023-07-27T05:00:00Z',
              'nextRecurrence': '2023-07-27T05:00:00Z',
              'nextContributionAmount': 750,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:42:11Z',
              'dateStarted': '2023-02-28T06:00:00Z',
            },
            {
              'spendingId': 134,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Freshbooks',
              'description': 'Every month on the 10th',
              'targetAmount': 1700,
              'currentAmount': 1700,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10',
              'lastRecurrence': '2023-07-10T05:00:00Z',
              'nextRecurrence': '2023-07-10T05:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:40:46Z',
              'dateStarted': '2023-03-10T06:00:00Z',
            },
            {
              'spendingId': 137,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Sentry',
              'description': 'Every month on the 25th',
              'targetAmount': 2900,
              'currentAmount': 1450,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=25',
              'lastRecurrence': '2023-07-25T05:00:00Z',
              'nextRecurrence': '2023-07-25T05:00:00Z',
              'nextContributionAmount': 1450,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:42:41Z',
              'dateStarted': '2023-03-25T05:00:00Z',
            },
            {
              'spendingId': 171,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'ngrok',
              'description': 'Every year on the 26th of June',
              'targetAmount': 6000,
              'currentAmount': 250,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=YEARLY;INTERVAL=1;BYMONTH=6;BYMONTHDAY=26',
              'lastRecurrence': '2024-06-26T05:00:00Z',
              'nextRecurrence': '2024-06-26T05:00:00Z',
              'nextContributionAmount': 250,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-06-28T15:59:10Z',
              'dateStarted': '2023-06-25T05:00:00Z',
            },
            {
              'spendingId': 191,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'Cloud Production',
              'description': 'Every month on the 1st',
              'targetAmount': 28000,
              'currentAmount': 28000,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-08-01T05:00:00Z',
              'nextRecurrence': '2023-08-01T05:00:00Z',
              'nextContributionAmount': 0,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2022-11-07T15:09:16Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            },
            {
              'spendingId': 135,
              'bankAccountId': 12,
              'fundingScheduleId': 3,
              'spendingType': 0,
              'name': 'G-Suite ($12)',
              'description': 'Every month on the 1st',
              'targetAmount': 1200,
              'currentAmount': 0,
              'usedAmount': 0,
              'recurrenceRule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
              'lastRecurrence': '2023-08-01T05:00:00Z',
              'nextRecurrence': '2023-08-01T05:00:00Z',
              'nextContributionAmount': 600,
              'isBehind': false,
              'isPaused': false,
              'dateCreated': '2021-12-14T16:41:18Z',
              'dateStarted': '2023-03-01T06:00:00Z',
            },
          ]));
        }),
        rest.get('/api/bank_accounts/12/funding_schedules', (_req, res, ctx) => {
          return res(ctx.json([
            {
              'fundingScheduleId': 3,
              'bankAccountId': 12,
              'name': 'Elliot\'s Contribution',
              'description': '15th and last day of every month',
              'rule': 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
              'excludeWeekends': true,
              'waitForDeposit': false,
              'estimatedDeposit': null,
              'lastOccurrence': '2023-06-30T05:00:00Z',
              'nextOccurrence': '2023-07-14T05:00:00Z',
              'dateStarted': '2023-02-28T06:00:00Z',
            },
          ]));
        }),
      ],
    },
  },
};

export default meta;

export const Transactions: StoryObj<typeof MonetrWrapper> = {
  name: 'Transactions',
  render: () => (
    <MonetrWrapper>
      <BankView>
        <TransactionsView />
      </BankView>
    </MonetrWrapper>
  ),
};

export const Expenses: StoryObj<typeof MonetrWrapper> = {
  name: 'Expenses',
  render: () => (
    <MonetrWrapper>
      <BankView>
        <ExpensesView />
      </BankView>
    </MonetrWrapper>
  ),
};

