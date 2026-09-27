-- Fixes issue where goals and expenses were not being rolled up properly because of how I was joining. Now the sums of
-- the goal and expense balances are determined and then joined.
CREATE
OR REPLACE VIEW "balances" AS (
SELECT "bank_account"."bank_account_id",
       "bank_account"."account_id",
       "bank_account"."current_balance"             AS "current",
       "bank_account"."available_balance"           AS "available",
       "bank_account"."available_balance" - SUM(COALESCE("expense"."current_amount", 0)) -
       SUM(COALESCE("goal"."current_amount", 0))    AS "safe",
       SUM(COALESCE("expense"."current_amount", 0)) AS "expenses",
       SUM(COALESCE("goal"."current_amount", 0))    AS "goals"
FROM "bank_accounts" AS "bank_account"
LEFT JOIN (
    SELECT spending.bank_account_id, spending.account_id, SUM(spending.current_amount) as "current_amount"
    FROM spending
    WHERE spending.spending_type = 0
    GROUP BY spending.bank_account_id, spending.account_id
) AS "expense"
ON "expense"."bank_account_id" = "bank_account"."bank_account_id" AND
   "expense"."account_id" = "bank_account"."account_id"
LEFT JOIN (
    SELECT spending.bank_account_id, spending.account_id, SUM(spending.current_amount) as "current_amount"
    FROM spending
    WHERE spending.spending_type = 1
   GROUP BY spending.bank_account_id, spending.account_id
) AS "goal" ON "goal"."bank_account_id" = "bank_account"."bank_account_id" AND
                                           "goal"."account_id" = "bank_account"."account_id"
GROUP BY "bank_account"."bank_account_id", "bank_account"."account_id"
);