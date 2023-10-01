ALTER TABLE "funding_schedules" ADD COLUMN "rruleset" TEXT;

WITH migrated_rules AS (
  SELECT
    funding_schedule_id,
    -- 20120201T023000Z
    'DTSTART:' || to_char(date_started AT TIME ZONE 'UTC', 'YYYYMMDD') || to_char(date_started AT TIME ZONE 'UTC', 'THH24MISSZ') || E'\n' ||
      'RRULE:' || rule AS "ruleset"
  FROM funding_schedules;
);

UPDATE "funding_schedules"
SET "ruleset"="migrated_rules"."ruleset"
FROM "migrated_rules"
WHERE "funding_schedules"."funding_schedule_id"="migrated_rules"."funding_schedule_id";

ALTER TABLE "funding_schedules" ALTER COLUMN "rruleset" SET NOT NULL;

