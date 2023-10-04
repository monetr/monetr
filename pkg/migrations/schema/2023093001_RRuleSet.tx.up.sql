--- Funding Schedules
ALTER TABLE "funding_schedules" ADD COLUMN "ruleset" TEXT;

WITH migrated_rules AS (
  SELECT
    funding_schedule_id,
    -- Needs to be done in two parts otherwise to_char doesn't handle the `T` separator properly.
    'DTSTART:' || to_char(date_started AT TIME ZONE 'UTC', 'YYYYMMDD') || to_char(date_started AT TIME ZONE 'UTC', 'THH24MISSZ') || E'\n' ||
      'RRULE:' || rule AS "ruleset"
  FROM funding_schedules
)
UPDATE "funding_schedules"
SET "ruleset"="migrated_rules"."ruleset"
FROM "migrated_rules"
WHERE "funding_schedules"."funding_schedule_id"="migrated_rules"."funding_schedule_id";

ALTER TABLE "funding_schedules" ALTER COLUMN "ruleset" SET NOT NULL;
ALTER TABLE "funding_schedules" DROP COLUMN "rule";
ALTER TABLE "funding_schedules" DROP COLUMN "date_started";

--- Spending
ALTER TABLE "spending" ADD COLUMN "ruleset" TEXT;

WITH migrated_rules AS (
  SELECT
    spending_id,
    -- Needs to be done in two parts otherwise to_char doesn't handle the `T` separator properly.
    'DTSTART:' || to_char(date_started AT TIME ZONE 'UTC', 'YYYYMMDD') || to_char(date_started AT TIME ZONE 'UTC', 'THH24MISSZ') || E'\n' ||
      'RRULE:' || recurrence_rule AS "ruleset"
  FROM spending
)
UPDATE "spending"
SET "ruleset"="migrated_rules"."ruleset"
FROM "migrated_rules"
WHERE "spending"."spending_id"="migrated_rules"."spending_id";
ALTER TABLE "spending" DROP COLUMN "recurrence_rule";
ALTER TABLE "spending" DROP COLUMN "date_started";
