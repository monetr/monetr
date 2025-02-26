### Authentication Endpoints (No auth required)

`POST /authentication/login` - Login
`GET /authentication/logout` - Logout
`POST /authentication/register` - Register new account
`POST /authentication/verify` - Verify email
`POST /authentication/verify/resend` - Resend verification
`POST /authentication/forgot` - Forgot password
`POST /authentication/reset` - Reset password
`POST /authentication/multifactor` - MFA verification

User Management (Auth required)

GET /users/me - Get current user info
PUT /users/security/password - Change password
POST /users/security/totp/setup - Setup TOTP (2FA)
POST /users/security/totp/confirm - Confirm TOTP setup
Billing (Auth required)

POST /billing/create_checkout - Create checkout session
GET /billing/checkout/:checkoutSessionId - Get checkout status
GET /billing/portal - Get billing portal
Bank & Account Management (Auth + Active subscription required)

GET /bank_accounts - List bank accounts
GET /bank_accounts/:bankAccountId - Get specific account
PUT /bank_accounts/:bankAccountId - Update account
GET /bank_accounts/:bankAccountId/balances - Get balances
POST /bank_accounts - Add bank account
Transactions

GET /bank_accounts/:bankAccountId/transactions - List transactions
GET /bank_accounts/:bankAccountId/transactions/:transactionId - Get transaction
POST /bank_accounts/:bankAccountId/transactions - Create transaction
PUT /bank_accounts/:bankAccountId/transactions/:transactionId - Update transaction
DELETE /bank_accounts/:bankAccountId/transactions/:transactionId - Delete transaction
Links & Plaid Integration

GET /links - List links
POST /links - Create link
PUT /links/:linkId - Update link
DELETE /links/:linkId - Delete link
PUT /plaid/link/update/:linkId - Update Plaid link
GET /plaid/link/token/new - Get new Plaid token
POST /plaid/link/token/callback - Plaid token callback
Funding & Spending

GET /bank_accounts/:bankAccountId/funding_schedules - List funding schedules
POST /bank_accounts/:bankAccountId/funding_schedules - Create funding schedule
GET /bank_accounts/:bankAccountId/spending - List spending
POST /bank_accounts/:bankAccountId/spending - Create spending
POST /bank_accounts/:bankAccountId/spending/transfer - Transfer spending
Forecasting

GET /bank_accounts/:bankAccountId/forecast - Get forecast
POST /bank_accounts/:bankAccountId/forecast/spending - Forecast new spending
POST /bank_accounts/:bankAccountId/forecast/next_funding - Forecast next funding
Miscellaneous

GET /config - Get configuration
POST /icons/search - Search icons
GET /locale/currency - List currencies
GET /institutions/:institutionId - Get institution details
Note: Most endpoints require:

Authentication (via cookie from login)
Active subscription (if billing is enabled)
Proper CSRF tokens (if enabled)
The API follows RESTful practices and returns appropriate HTTP status codes. Error responses will include an error field with a description of what went wrong.