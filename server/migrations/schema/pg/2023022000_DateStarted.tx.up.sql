ALTER TABLE "spending" ADD COLUMN "date_started" TIMESTAMPTZ NULL;
UPDATE "spending" SET "date_started" = "next_recurrence" WHERE "date_started" IS NULL;
ALTER TABLE "spending" ALTER COLUMN "date_started" SET NOT NULL;

ALTER TABLE "funding_schedules" ADD COLUMN "date_started" TIMESTAMPTZ NULL;
UPDATE "funding_schedules" SET "date_started" = "next_occurrence" WHERE "date_started" IS NULL;
ALTER TABLE "funding_schedules" ALTER COLUMN "date_started" SET NOT NULL;
