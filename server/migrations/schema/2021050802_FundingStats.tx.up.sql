CREATE VIEW "funding_stats" AS
(
SELECT "bank_account"."bank_account_id",
       "bank_account"."account_id",
       "funding_schedule"."funding_schedule_id",
       COUNT("expenses"."spending_id")                         AS "number_of_expenses",
       COUNT("goals"."spending_id")                            AS "number_of_goals",
       SUM(COALESCE("expenses"."next_contribution_amount", 0)) AS "next_expense_contribution",
       SUM(COALESCE("goals"."next_contribution_amount", 0))    AS "next_goal_contribution"
FROM "bank_accounts" AS "bank_account"
         INNER JOIN "funding_schedules" AS "funding_schedule" ON
        "funding_schedule"."account_id" = "bank_account"."account_id" AND
        "funding_schedule"."bank_account_id" = "bank_account"."bank_account_id"
         LEFT JOIN "spending" AS "expenses" ON
        "expenses"."account_id" = "funding_schedule"."account_id" AND
        "expenses"."bank_account_id" = "funding_schedule"."bank_account_id" AND
        "expenses"."funding_schedule_id" = "funding_schedule"."funding_schedule_id" AND
        "expenses"."spending_type" = 0
         LEFT JOIN "spending" AS "goals" ON
        "goals"."account_id" = "funding_schedule"."account_id" AND
        "goals"."bank_account_id" = "funding_schedule"."bank_account_id" AND
        "goals"."funding_schedule_id" = "funding_schedule"."funding_schedule_id" AND
        "goals"."spending_type" = 1
GROUP BY "bank_account"."bank_account_id", "bank_account"."account_id", "funding_schedule"."funding_schedule_id"
    );