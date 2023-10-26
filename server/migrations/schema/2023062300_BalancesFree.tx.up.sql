-- Replace "safe" with "free"
DROP VIEW IF EXISTS "balances";
-- You can't rename a column in a view, so you have to drop the view first then recreate it. You cannot create or
-- replace a column name change.
-- https://dba.stackexchange.com/questions/586/cant-rename-columns-in-postgresql-views-with-create-or-replace
CREATE VIEW "balances" AS (
  SELECT
    "bank_account"."bank_account_id",
    "bank_account"."account_id",
    "bank_account"."current_balance"   AS "current",
    "bank_account"."available_balance" AS "available",
    "bank_account"."available_balance" - SUM(COALESCE("expense"."current_amount", 0)) - SUM(COALESCE("goal"."current_amount", 0)) AS "free",
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
