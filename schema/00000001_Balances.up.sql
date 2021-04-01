DROP VIEW "balances";
CREATE VIEW "balances" AS
(
SELECT "bank_account"."bank_account_id",
       "bank_account"."account_id",
       "bank_account"."current_balance"             AS "current",
       "bank_account"."available_balance"           AS "available",
       "bank_account"."available_balance" - SUM(COALESCE("expense"."current_amount", 0)) -
       SUM(COALESCE("goal"."current_amount", 0))       AS "safe",
       SUM(COALESCE("expense"."current_amount", 0)) AS "expenses",
       SUM(COALESCE("goal"."current_amount", 0))    AS "goals"
FROM "bank_accounts" AS "bank_account"
         LEFT JOIN "spending" AS "expense"
                   ON "expense"."bank_account_id" = "bank_account"."bank_account_id" AND
                      "expense"."account_id" = "bank_account"."account_id" AND
                      "expense"."spending_type" = 0 -- 0 means its an expense
         LEFT JOIN "spending" AS "goal" ON "goal"."bank_account_id" = "bank_account"."bank_account_id" AND
                                           "goal"."account_id" = "bank_account"."account_id" AND
                                           "goal"."spending_type" = 1 -- 1 means its a goal
GROUP BY "bank_account"."bank_account_id", "bank_account"."account_id"
    );