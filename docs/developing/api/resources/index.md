# REST API Resources

Available resources for the monetr REST API.

| Resource                                  | Available endpoints                                                                      |
|:------------------------------------------|:-----------------------------------------------------------------------------------------|
| [Authentication](authentication.md)       | `/api/authenication/login`, `/api/authentication/logout`, `/api/authentication/register` |
| [Bank Accounts](bank_accounts.md)         | `/api/bank_accounts`                                                                     |
| [Funding Schedules](funding_schedules.md) | `/api/bank_accounts/:bankAccountId/funding_schedules`                                    |
| [Links](links.md)                         | `/api/links`                                                                             |
| [Plaid Links](plaid_links.md)             | `/api/plaid/token/new`, `/api/plaid/token/callback`                                      |
| [Spending](spending.md)                   | `/api/bank_accounts/:bankAccountId/spending`                                             |
| [Transactions](transactions.md)           | `/api/bank_accounts/:bankAccountId/transactions`                                         |
| [User](user.md)                           | `/api/users/me`                                                                          |
