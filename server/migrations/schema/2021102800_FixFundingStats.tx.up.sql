DROP VIEW "funding_stats";

CREATE VIEW "funding_stats" AS
(
SELECT "bank_account"."bank_account_id",
       "bank_account"."account_id",
       "funding_schedule"."funding_schedule_id",
       (SELECT COALESCE(SUM("expenses"."next_contribution_amount"), 0)
        FROM "spending" AS "expenses"
        WHERE "expenses"."account_id" = "funding_schedule"."account_id"
          AND "expenses"."bank_account_id" = "funding_schedule"."bank_account_id"
          AND "expenses"."funding_schedule_id" = "funding_schedule"."funding_schedule_id"
          AND "expenses"."spending_type" = 0) AS "next_expense_contribution",
       (SELECT COALESCE(SUM("goals"."next_contribution_amount"), 0)
        FROM "spending" AS "goals"
        WHERE "goals"."account_id" = "funding_schedule"."account_id"
          AND "goals"."bank_account_id" = "funding_schedule"."bank_account_id"
          AND "goals"."funding_schedule_id" = "funding_schedule"."funding_schedule_id"
          AND "goals"."spending_type" = 1)    AS "next_goal_contribution"
FROM "bank_accounts" AS "bank_account"
         INNER JOIN "funding_schedules" AS "funding_schedule" ON
            "funding_schedule"."account_id" = "bank_account"."account_id" AND
            "funding_schedule"."bank_account_id" = "bank_account"."bank_account_id"
GROUP BY "bank_account"."bank_account_id",
         "bank_account"."account_id",
         "funding_schedule"."bank_account_id",
         "funding_schedule"."account_id",
         "funding_schedule"."funding_schedule_id"
    );