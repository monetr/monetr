/* eslint-disable max-len */
import MockAdapter from 'axios-mock-adapter';

// apiSampleResponses will fill the mock adapter with most of the API calls that the application needs to render most
// pages. Some pages may not be able to render with this data and additional mocks may be needed. This should only be
// called once per test.
export default function apiSampleResponses(mockAxios: MockAdapter) {
  mockAxios.onGet('/api/config').reply(200, {
    'requireLegalName': true,
    'requirePhoneNumber': true,
    'verifyLogin': false,
    'verifyRegister': false,
    'verifyEmailAddress': true,
    'verifyForgotPassword': false,
    'allowSignUp': true,
    'allowForgotPassword': true,
    'longPollPlaidSetup': true,
    'requireBetaCode': true,
    'initialPlan': {
      'price': 199,
    },
    'billingEnabled': true,
    'iconsEnabled': true,
    'plaidEnabled': true,
    'manualEnabled': false,
    'release': '0.17.16',
    'revision': '8df5505b7e5273f061d90ddf19e4c1cfca2b4f4f',
    'buildType': 'binary',
    'buildTime': '2024-08-28T02:42:56Z',
  });
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
  mockAxios.onGet('/api/links').reply(200, [
    {
      'linkId': 'link_01gds6eqsqacg48p0azb3wcpsq',
      'linkType': 1,
      'plaidLink': {
        'products': [
          'transactions',
        ],
        'status': 2,
        'expirationDate': null,
        'newAccountsAvailable': false,
        'institutionId': 'ins_116794',
        'institutionName': 'Mercury',
        'lastManualSync': '2024-07-06T12:59:09.51222Z',
        'lastSuccessfulUpdate': '2024-08-29T12:00:01.176597Z',
        'lastAttemptedUpdate': '2024-08-29T12:00:01.17665Z',
        'updatedAt': '2024-03-19T06:17:32.335106Z',
        'createdAt': '2022-09-25T02:08:40.758642Z',
        'createdBy': 'user_01hym36e8ewaq0hxssb1m3k4ha',
      },
      'institutionName': 'Mercury',
      'description': null,
      'createdAt': '2022-09-25T02:08:40.758642Z',
      'createdBy': 'user_01hym36e8ewaq0hxssb1m3k4ha',
      'updatedAt': '2024-03-19T06:17:32.335106Z',
      'deletedAt': null,
    },
  ]);
  mockAxios.onGet('/api/institutions/ins_116794').reply(200, {
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
  });
  mockAxios.onGet('/api/bank_accounts').reply(200, [
    {
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
      'currency': 'USD',
      'lastUpdated': '2024-08-27T08:53:48.555368Z',
      'createdAt': '2022-09-25T02:08:40.758642Z',
      'updatedAt': '2024-03-19T06:17:32.335106Z',
    },
  ]);
  mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb').reply(200, {
    'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
    'linkId': 'link_01gds6eqsqacg48p0azb3wcpsq',
    'plaidBankAccount': {
      'name': 'Mercury Checking',
      'officialName': 'Mercury Checking',
      'mask': '2982',
      'availableBalance': 47986,
      'currentBalance': 47986,
      'limitBalance': null,
      'createdAt': '2024-03-19T15:15:10.31132Z',
      'createdBy': 'user_01hym36e8ewaq0hxssb1m3k4ha',
    },
    'availableBalance': 47986,
    'currentBalance': 47986,
    'mask': '2982',
    'name': 'Mercury Checking',
    'originalName': 'Mercury Checking',
    'accountType': 'depository',
    'accountSubType': 'checking',
    'status': 'active',
    'currency': 'USD',
    'lastUpdated': '2024-08-27T08:53:48.555368Z',
    'createdAt': '2022-09-25T02:08:40.758642Z',
    'updatedAt': '2024-03-19T06:17:32.335106Z',
  });
  mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules').reply(200, [
    {
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'name': 'Elliot\'s Contribution',
      'description': '15th and last day of every month',
      'ruleset': 'DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
      'excludeWeekends': true,
      'waitForDeposit': false,
      'estimatedDeposit': 22000,
      'lastRecurrence': '2024-08-15T05:00:00Z',
      'nextRecurrence': '2024-08-30T05:00:00Z',
      'nextRecurrenceOriginal': '2024-08-31T05:00:00Z',
    },
  ]);
  mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/funding_schedules/fund_01hym37k3kj4ghv67nfx7vkvr0').reply(200, {
    'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
    'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
    'name': 'Elliot\'s Contribution',
    'description': '15th and last day of every month',
    'ruleset': 'DTSTART:20230228T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
    'excludeWeekends': true,
    'waitForDeposit': false,
    'estimatedDeposit': 22000,
    'lastRecurrence': '2024-08-15T05:00:00Z',
    'nextRecurrence': '2024-08-30T05:00:00Z',
    'nextRecurrenceOriginal': '2024-08-31T05:00:00Z',
  });
  mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/balances').reply(200, {
    'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
    'current': 47986,
    'available': 47986,
    'free': 7724,
    'expenses': 30262,
    'goals': 10000,
  });
  mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/spending').reply(200, [
    {
      'spendingId': 'spnd_01h264znvxkghmxp5s0wmdbr9j',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'GitLab',
      'description': 'Every year on the 16th of June',
      'targetAmount': 34800,
      'currentAmount': 5800,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230606T050000Z\nRRULE:FREQ=YEARLY;INTERVAL=1;BYMONTH=6;BYMONTHDAY=16',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-06-16T05:00:00Z',
      'nextRecurrence': '2025-06-16T05:00:00Z',
      'nextContributionAmount': 1450,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2023-06-05T16:07:02.780623Z',
    },
    {
      'spendingId': 'spnd_01fpwx2qng9gbt32kdh2mdk7df',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'Freshbooks',
      'targetAmount': 1900,
      'currentAmount': 950,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230310T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-08-10T06:00:00Z',
      'nextRecurrence': '2024-09-10T06:00:00Z',
      'nextContributionAmount': 950,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2021-12-14T16:40:46Z',
    },
    {
      'spendingId': 'spnd_01gk23r0mgcekdf8bzkvf55f26',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 1,
      'name': 'Rainy Day',
      'targetAmount': 100,
      'currentAmount': 17724,
      'usedAmount': 58900,
      'ruleset': null,
      'lastSpentFrom': null,
      'lastRecurrence': '2023-12-31T06:00:00Z',
      'nextRecurrence': '2023-12-31T06:00:00Z',
      'nextContributionAmount': 0,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2022-11-29T16:32:58Z',
    },
    {
      'spendingId': 'spnd_01ggwaxcq0y6eazf56a4sq3f02',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'Google Voice',
      'targetAmount': 1292,
      'currentAmount': 646,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230301T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-08-01T06:00:00Z',
      'nextRecurrence': '2024-09-01T06:00:00Z',
      'nextContributionAmount': 646,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2022-11-02T14:11:24Z',
    },
    {
      'spendingId': 'spnd_01g6ngke5g9mgrn9kt7pjdcqr8',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'ngrok',
      'description': 'Every year on the 26th of June',
      'targetAmount': 6000,
      'currentAmount': 780,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230625T050000Z\nRRULE:FREQ=YEARLY;INTERVAL=1;BYMONTH=6;BYMONTHDAY=26',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-06-26T05:00:00Z',
      'nextRecurrence': '2025-06-26T05:00:00Z',
      'nextContributionAmount': 261,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2022-06-28T15:59:10Z',
    },
    {
      'spendingId': 'spnd_01h0jgggcdng3ztx90zwh6nb6f',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'Domains',
      'description': 'Every year on the 1st of April',
      'targetAmount': 14000,
      'currentAmount': 2331,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20240401T050000Z\nRRULE:FREQ=YEARLY;INTERVAL=1;BYMONTH=4;BYMONTHDAY=1',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-04-01T05:00:00Z',
      'nextRecurrence': '2025-04-01T05:00:00Z',
      'nextContributionAmount': 777,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2023-05-16T14:47:58.09301Z',
    },
    {
      'spendingId': 'spnd_01h0dxnp2nm9q90pdqemp1z5m1',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'Plaid',
      'description': 'Every month on the 10th',
      'targetAmount': 500,
      'currentAmount': 365,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230610T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-08-10T05:00:00Z',
      'nextRecurrence': '2024-09-10T05:00:00Z',
      'nextContributionAmount': 135,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2023-05-14T20:01:47.09268Z',
    },
    {
      'spendingId': 'spnd_01fpwx3pxg6yaspkx0wn3fptkf',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'Google G-Suite',
      'targetAmount': 1440,
      'currentAmount': 720,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230301T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-08-01T06:00:00Z',
      'nextRecurrence': '2024-09-01T06:00:00Z',
      'nextContributionAmount': 720,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2021-12-14T16:41:18Z',
    },
    {
      'spendingId': 'spnd_01gh9a6yb0weejnxk5y0e66s7s',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'Google Cloud Production',
      'targetAmount': 25000,
      'currentAmount': 12588,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230301T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-08-01T06:00:00Z',
      'nextRecurrence': '2024-09-01T06:00:00Z',
      'nextContributionAmount': 12412,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2022-11-07T15:09:16Z',
    },
    {
      'spendingId': 'spnd_01hcjv04tytz5xatf3dhers5kg',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'P.O. Box',
      'description': 'Every year on the 13th of October',
      'targetAmount': 7000,
      'currentAmount': 6082,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20231013T050000Z\nRRULE:FREQ=YEARLY;INTERVAL=1;BYMONTH=10;BYMONTHDAY=13',
      'lastSpentFrom': null,
      'lastRecurrence': '2023-10-13T05:00:00Z',
      'nextRecurrence': '2024-10-13T05:00:00Z',
      'nextContributionAmount': 306,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2023-10-12T20:59:38.206294Z',
    },
    {
      'spendingId': 'spnd_01fpwx6ye0kehpverathqakm9x',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'GitHub',
      'targetAmount': 800,
      'currentAmount': 0,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230319T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=19',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-08-19T05:00:00Z',
      'nextRecurrence': '2024-09-19T05:00:00Z',
      'nextContributionAmount': 400,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2021-12-14T16:43:04Z',
    },
    {
      'spendingId': 'spnd_01fpwx67z8djhb8bqcy17mrzfe',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'fundingScheduleId': 'fund_01hym37k3kj4ghv67nfx7vkvr0',
      'spendingType': 0,
      'name': 'Sentry',
      'targetAmount': 3446,
      'currentAmount': 0,
      'usedAmount': 0,
      'ruleset': 'DTSTART:20230325T050000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=25',
      'lastSpentFrom': null,
      'lastRecurrence': '2024-08-25T05:00:00Z',
      'nextRecurrence': '2024-09-25T05:00:00Z',
      'nextContributionAmount': 1723,
      'isBehind': false,
      'isPaused': false,
      'createdAt': '2021-12-14T16:42:41Z',
    },
  ]);
  mockAxios.onGet('/api/bank_accounts/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions').reply(200, [
    {
      'transactionId': 'txn_01j68vszqeq30t7jz7atk9yd9r',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 3446,
      'spendingId': 'spnd_01fpwx67z8djhb8bqcy17mrzfe',
      'spendingAmount': 2900,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'FOOD_AND_DRINK_GROCERIES',
      'date': '2024-08-27T05:00:00Z',
      'name': 'Sentry.',
      'originalName': 'SENTRY. Merchant name: Sentry',
      'merchantName': 'Sentry.',
      'originalMerchantName': 'Sentry.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-27T02:49:28.059Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j5m4enxf50k3d2h2c9dtst4z',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 803,
      'spendingId': 'spnd_01fpwx6ye0kehpverathqakm9x',
      'spendingAmount': 800,
      'categories': [
        'Service',
        'Computers',
        'Software Development',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-08-19T05:00:00Z',
      'name': 'GitHub',
      'originalName': 'GITHUB, INC.. Merchant name: Github',
      'merchantName': 'GitHub',
      'originalMerchantName': 'GitHub',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-19T01:36:31.663Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j59tnczpyqwq7sxw4gzbmkvx',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 270,
      'spendingId': 'spnd_01h0dxnp2nm9q90pdqemp1z5m1',
      'spendingAmount': 270,
      'categories': [
        'Service',
        'Business Services',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-08-16T05:00:00Z',
      'name': 'Plaid',
      'originalName': 'Plaid',
      'merchantName': 'Plaid Technologies Inc.',
      'originalMerchantName': 'Plaid Technologies Inc.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-15T01:33:01.814Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j5b3p0waxq4894dgteb58s6e',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': -22000,
      'spendingId': null,
      'categories': [
        'Transfer',
        'Payroll',
      ],
      'category': 'INCOME_WAGES',
      'date': '2024-08-15T05:00:00Z',
      'name': 'Elliot\'s Contribution',
      'originalName': '5176 TREASURY PR; DEPOSIT; COURANT ELLIOT. Merchant name: 5176 TREASURY PR',
      'originalMerchantName': '',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-15T13:29:53.802Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j4vn659athnneb4jd6yf2twa',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 1900,
      'spendingId': 'spnd_01fpwx2qng9gbt32kdh2mdk7df',
      'spendingAmount': 1900,
      'categories': [
        'Service',
        'Financial',
        'Accounting and Bookkeeping',
      ],
      'category': 'GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING',
      'date': '2024-08-10T05:00:00Z',
      'name': 'FreshBooks',
      'originalName': 'FRESHBOOKS. Merchant name: Freshbooks',
      'merchantName': 'FreshBooks',
      'originalMerchantName': 'FreshBooks',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-09T13:27:57.482Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j47xct43472gtrv6dkp8s9s8',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 1440,
      'spendingId': 'spnd_01fpwx3pxg6yaspkx0wn3fptkf',
      'spendingAmount': 1440,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-08-02T05:00:00Z',
      'name': 'Gsuite Monetr.a.',
      'originalName': 'GSUITE MONETR.A. Merchant name: Gsuite',
      'merchantName': 'Gsuite Monetr.a.',
      'originalMerchantName': 'Gsuite Monetr.a.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-01T21:26:35.396Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j47xct4230fpdj4gt8trf021',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 24824,
      'spendingId': 'spnd_01gh9a6yb0weejnxk5y0e66s7s',
      'spendingAmount': 24824,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE',
      'date': '2024-08-02T05:00:00Z',
      'name': 'Cloud',
      'originalName': 'Google CLOUD 2xHT9W. Merchant name: Google Cloud',
      'merchantName': 'Cloud',
      'originalMerchantName': 'Cloud',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-01T21:26:35.396Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j47xct40x2vwfhst88k4j9zs',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 1294,
      'spendingId': 'spnd_01ggwaxcq0y6eazf56a4sq3f02',
      'spendingAmount': 1292,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'TRANSFER_OUT_ACCOUNT_TRANSFER',
      'date': '2024-08-02T05:00:00Z',
      'name': 'Svcsmonetr.app.',
      'originalName': 'SVCSMONETR.APP. Merchant name: Google',
      'merchantName': 'Svcsmonetr.app.',
      'originalMerchantName': 'Svcsmonetr.app.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-08-01T21:26:35.396Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j45asp0shxbda8abgmqsa6mz',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': -22000,
      'spendingId': null,
      'categories': [
        'Transfer',
        'Payroll',
      ],
      'category': 'INCOME_WAGES',
      'date': '2024-07-31T05:00:00Z',
      'name': 'Elliot\'s Contribution',
      'originalName': '5176 TREASURY PR; DEPOSIT; COURANT ELLIOT. Merchant name: 5176 TREASURY PR',
      'originalMerchantName': '',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-31T21:23:05.369Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j3s2r562navpr3y1y27rn95j',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 3914,
      'spendingId': 'spnd_01fpwx67z8djhb8bqcy17mrzfe',
      'spendingAmount': 2900,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'FOOD_AND_DRINK_GROCERIES',
      'date': '2024-07-27T05:00:00Z',
      'name': 'Sentry.',
      'originalName': 'SENTRY. Merchant name: Sentry',
      'merchantName': 'Sentry.',
      'originalMerchantName': 'Sentry.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-27T03:11:33.57Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j33sp41a5b2s3nnkddjw1grc',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 800,
      'spendingId': 'spnd_01fpwx6ye0kehpverathqakm9x',
      'spendingAmount': 800,
      'categories': [
        'Service',
        'Computers',
        'Software Development',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-07-19T05:00:00Z',
      'name': 'GitHub',
      'originalName': 'GITHUB, INC.. Merchant name: Github',
      'merchantName': 'GitHub',
      'originalMerchantName': 'GitHub',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-18T20:49:06.602Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j2r4mbwb1d56d5be9h5y6jxq',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 210,
      'spendingId': 'spnd_01h0dxnp2nm9q90pdqemp1z5m1',
      'spendingAmount': 210,
      'categories': [
        'Service',
        'Business Services',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-07-16T05:00:00Z',
      'name': 'Plaid',
      'originalName': 'Plaid',
      'merchantName': 'Plaid Technologies Inc.',
      'originalMerchantName': 'Plaid Technologies Inc.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-14T08:09:30.251Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j2vc5pkhbmjp7kmj7ede9vbe',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': -22000,
      'spendingId': null,
      'categories': [
        'Transfer',
        'Payroll',
      ],
      'category': 'INCOME_WAGES',
      'date': '2024-07-15T05:00:00Z',
      'name': 'Elliot\'s Contribution',
      'originalName': '5176 TREASURY PR; DEPOSIT; COURANT ELLIOT. Merchant name: 5176 TREASURY PR',
      'originalMerchantName': '',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-15T14:19:01.617Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j2bt5rjem60dgvf6ymm8zrvw',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 1900,
      'spendingId': 'spnd_01fpwx2qng9gbt32kdh2mdk7df',
      'spendingAmount': 1900,
      'categories': [
        'Service',
        'Financial',
        'Accounting and Bookkeeping',
      ],
      'category': 'GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING',
      'date': '2024-07-10T05:00:00Z',
      'name': 'FreshBooks',
      'originalName': 'FRESHBOOKS. Merchant name: Freshbooks',
      'merchantName': 'FreshBooks',
      'originalMerchantName': 'FreshBooks',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-09T13:15:52.782Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j27xtf159hx1j8ajzev4whtb',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 4000,
      'spendingId': 'spnd_01h0jgggcdng3ztx90zwh6nb6f',
      'spendingAmount': 3040,
      'categories': [
        'Transfer',
        'Debit',
      ],
      'category': 'TRANSFER_OUT_ACCOUNT_TRANSFER',
      'date': '2024-07-08T05:00:00Z',
      'name': 'Square Space Domains',
      'originalName': 'SQSP* INV140023242. Merchant name: Sqsp',
      'merchantName': 'INV',
      'originalMerchantName': 'INV',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-08T01:02:39.141Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j1tzj5snpm0c2zx47f1bma2d',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 6000,
      'spendingId': 'spnd_01g6ngke5g9mgrn9kt7pjdcqr8',
      'spendingAmount': 6000,
      'categories': [
        'Transfer',
        'Debit',
      ],
      'category': 'FOOD_AND_DRINK_GROCERIES',
      'date': '2024-07-03T05:00:00Z',
      'name': 'Ngrok',
      'originalName': 'NGROK 6RWWW7NJFQD-0004. Merchant name: Ngrok 6rwww7njfqd 0004',
      'merchantName': 'Ngrok',
      'originalMerchantName': 'Ngrok',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-03T00:22:57.078Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j1qqt56ngf8paqzp4k8vwnvq',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 20775,
      'spendingId': 'spnd_01gh9a6yb0weejnxk5y0e66s7s',
      'spendingAmount': 20775,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-07-02T05:00:00Z',
      'name': 'Cloud Vb',
      'originalName': 'Google CLOUD vB57xH. Merchant name: Google Cloud',
      'merchantName': 'Cloud Vb',
      'originalMerchantName': 'Cloud Vb',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-01T18:09:46.709Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j1qqt56k7ytp6ynqz7q82yzn',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 1287,
      'spendingId': 'spnd_01ggwaxcq0y6eazf56a4sq3f02',
      'spendingAmount': 1287,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'TRANSFER_OUT_ACCOUNT_TRANSFER',
      'date': '2024-07-02T05:00:00Z',
      'name': 'Svcsmonetr.app.',
      'originalName': 'SVCSMONETR.APP. Merchant name: Google',
      'merchantName': 'Svcsmonetr.app.',
      'originalMerchantName': 'Svcsmonetr.app.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-01T18:09:46.709Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j1q2sqwwbmvecxd50t453x40',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 1440,
      'spendingId': 'spnd_01fpwx3pxg6yaspkx0wn3fptkf',
      'spendingAmount': 1440,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-07-02T05:00:00Z',
      'name': 'Gsuite Monetr.a.',
      'originalName': 'GSUITE MONETR.A. Merchant name: Gsuite',
      'merchantName': 'Gsuite Monetr.a.',
      'originalMerchantName': 'Gsuite Monetr.a.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-07-01T12:02:32.988Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j1fzxja9mcy0tj6wb62b5n0p',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': -22000,
      'spendingId': null,
      'categories': [
        'Transfer',
        'Payroll',
      ],
      'category': 'INCOME_WAGES',
      'date': '2024-06-28T05:00:00Z',
      'name': 'Elliot\'s Contribution',
      'originalName': '5176 TREASURY PR; DEPOSIT; COURANT ELLIOT. Merchant name: 5176 TREASURY PR',
      'originalMerchantName': '',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-06-28T17:57:31.593Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j1e1d9p7qb41prh6jne7dtw7',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 34800,
      'spendingId': 'spnd_01h264znvxkghmxp5s0wmdbr9j',
      'spendingAmount': 34800,
      'categories': [
        'Service',
      ],
      'category': 'TRANSFER_OUT_ACCOUNT_TRANSFER',
      'date': '2024-06-28T05:00:00Z',
      'name': 'Gitlab Inc..',
      'originalName': 'GITLAB INC.. Merchant name: Gitlab',
      'merchantName': 'Gitlab Inc..',
      'originalMerchantName': 'Gitlab Inc..',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-06-27T23:45:06.759Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j1c2cq8dwdpjyvs8fxpkeh5q',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 2900,
      'spendingId': 'spnd_01fpwx67z8djhb8bqcy17mrzfe',
      'spendingAmount': 2900,
      'categories': [
        'Shops',
        'Digital Purchase',
      ],
      'category': 'FOOD_AND_DRINK_GROCERIES',
      'date': '2024-06-27T05:00:00Z',
      'name': 'Sentry.',
      'originalName': 'SENTRY. Merchant name: Sentry',
      'merchantName': 'Sentry.',
      'originalMerchantName': 'Sentry.',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-06-27T05:23:47.597Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j0psb1fgb1zcc696h7psm9c8',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': 800,
      'spendingId': 'spnd_01fpwx6ye0kehpverathqakm9x',
      'spendingAmount': 800,
      'categories': [
        'Service',
        'Computers',
        'Software Development',
      ],
      'category': 'GENERAL_SERVICES_OTHER_GENERAL_SERVICES',
      'date': '2024-06-19T05:00:00Z',
      'name': 'GitHub',
      'originalName': 'GITHUB, INC.. Merchant name: Github',
      'merchantName': 'GitHub',
      'originalMerchantName': 'GitHub',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-06-18T23:01:32.273Z',
      'deletedAt': null,
    },
    {
      'transactionId': 'txn_01j0cfj0ajsrqx2c0b41ewyhnj',
      'bankAccountId': 'bac_01gds6eqsq7h5mgevwtmw3cyxb',
      'plaidTransaction': null,
      'pendingPlaidTransaction': null,
      'amount': -20000,
      'spendingId': null,
      'categories': [
        'Transfer',
        'Payroll',
      ],
      'category': 'INCOME_WAGES',
      'date': '2024-06-14T05:00:00Z',
      'name': 'Elliot\'s Contribution',
      'originalName': '5176 TREASURY PR; DEPOSIT; COURANT ELLIOT. Merchant name: 5176 TREASURY PR',
      'originalMerchantName': '',
      'currency': 'USD',
      'isPending': false,
      'uploadIdentifier': null,
      'createdAt': '2024-06-14T22:58:10.386Z',
      'deletedAt': null,
    },
  ]);
}
