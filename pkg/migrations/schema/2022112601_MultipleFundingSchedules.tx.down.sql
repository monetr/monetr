ALTER TABLE "spending" ADD COLUMN "next_contribution_amount" BIGINT NULL; -- NULLABLE for now
ALTER TABLE "spending" ADD COLUMN "funding_schedule_id" BIGINT NULL; -- NULLABLE for now

-- NOTE: This DOES NOT WORK, if you already have spending objects with multiple funding schedules.
UPDATE "spending"
SET "spending"."next_contribution_amount" = "funding"."next_contribution_amount",
    "spending"."funding_schedule_id" = "funding"."funding_schedule_id"
FROM "spending" AS "spending"
INNER JOIN "spending_funding" AS "funding" ON "funding"."spending_id" = "spending"."spending_id";

ALTER TABLE "spending"
ADD CONSTRAINT "fk_spending_funding_schedules_funding_schedule_id_account_id_bank_account_id"
FOREIGN KEY ("funding_schedule_id", "account_id", "bank_account_id")
REFERENCES "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id");

UPDATE "spending" SET "next_contribution_amount" = 0 WHERE "next_contribution_amount" IS NULL;
ALTER TABLE "spending" ALTER COLUMN "next_contribution_amount" SET NOT NULL;
-- This will also break if the funding stuff was not migrated properly.
ALTER TABLE "spending" ALTER COLUMN "funding_schedule_id" SET NOT NULL;

DROP TABLE IF EXISTS "spending_funding";
