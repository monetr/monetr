ALTER TABLE "funding_schedules" ADD COLUMN "next_occurrence_original" TIMESTAMPTZ;
UPDATE "funding_schedules" SET "next_occurrence_original" = "next_occurrence";
ALTER TABLE "funding_schedules" ALTER COLUMN "next_occurrence_original" SET NOT NULL;
