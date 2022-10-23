# Links

Links are used to represent a connection between monetr and a bank. Links encapsulate multiple 
[bank accounts](../bank_accounts/index.md).

Links are primarily established via [Plaid](https://plaid.com){:target="_blank"}. But in the near future we want to 
support adding manual links that would allow the user to create bank accounts and transactions manually without 
requiring direct access to financial data via Plaid.

## Manual Links

Manual links are currently in development, progress can be tracked here:
[feature: Manual Links](https://github.com/monetr/monetr/milestone/6)

## Plaid Links

Links established via Plaid are managed automatically. As Plaid has more data available in the form of new 
transactions or updated balances, monetr will be notified and reflect that. This is what currently drives all of 
monetr's financial data for budgeting.

??? info
    
    If you have lost access to your monetr account, please contact support. If you want or need to revoke access to 
    your financial data that was accessed via Plaid. You can do so here: [my.plaid.com](https://my.plaid.com/)
