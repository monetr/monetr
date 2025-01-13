ALTER TABLE "bank_accounts" ADD COLUMN "currency" TEXT NOT NULL DEFAULT 'USD';
ALTER TABLE "plaid_bank_accounts" ADD COLUMN "currency" TEXT NOT NULL DEFAULT 'USD';

DROP VIEW "balances";
CREATE VIEW balances AS
 SELECT 
    bank_account.bank_account_id,
    bank_account.account_id,
    bank_account.currency,
    bank_account.current_balance AS current,
    bank_account.available_balance AS available,
    bank_account.limit_balance AS limit,
    bank_account.available_balance::numeric - sum(COALESCE(expense.current_amount, 0::numeric)) - sum(COALESCE(goal.current_amount, 0::numeric)) AS free,
    sum(COALESCE(expense.current_amount, 0::numeric)) AS expenses,
    sum(COALESCE(goal.current_amount, 0::numeric)) AS goals
   FROM bank_accounts bank_account
     LEFT JOIN ( 
       SELECT spending.bank_account_id,
              spending.account_id,
              sum(spending.current_amount) AS current_amount
       FROM spending
       WHERE spending.spending_type = 0 -- 0 = expenses
       GROUP BY spending.bank_account_id, spending.account_id
     ) expense ON expense.bank_account_id = bank_account.bank_account_id AND expense.account_id = bank_account.account_id
     LEFT JOIN (
       SELECT spending.bank_account_id,
              spending.account_id,
              sum(spending.current_amount) AS current_amount
       FROM spending
       WHERE spending.spending_type = 1 -- 1 is goals
       GROUP BY spending.bank_account_id, spending.account_id
     ) goal ON goal.bank_account_id = bank_account.bank_account_id AND goal.account_id = bank_account.account_id
  GROUP BY bank_account.bank_account_id, bank_account.account_id;
