ALTER TABLE "spending" ADD COLUMN IF NOT EXISTS "transaction_recurring_id" VARCHAR(32);
ALTER TABLE "funding_schedules" ADD COLUMN IF NOT EXISTS "transaction_recurring_id" VARCHAR(32);

UPDATE "spending"
SET "transaction_recurring_id" = "recurring"."transaction_recurring_id"
FROM "transaction_recurring" AS "recurring"
WHERE "recurring"."account_id" = "spending"."account_id"
  AND "recurring"."bank_account_id" = "spending"."bank_account_id"
  AND "recurring"."spending_id" = "spending"."spending_id";

UPDATE "funding_schedules"
SET "transaction_recurring_id" = "recurring"."transaction_recurring_id"
FROM "transaction_recurring" AS "recurring"
WHERE "recurring"."account_id" = "funding_schedules"."account_id"
  AND "recurring"."bank_account_id" = "funding_schedules"."bank_account_id"
  AND "recurring"."funding_schedule_id" = "funding_schedules"."funding_schedule_id";

ALTER TABLE "spending" ADD CONSTRAINT "fk_spending_transaction_recurring"
FOREIGN KEY ("transaction_recurring_id", "account_id", "bank_account_id")
REFERENCES "transaction_recurring" ("transaction_recurring_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("transaction_recurring_id");

ALTER TABLE "spending" ADD CONSTRAINT "uq_spending_transaction_recurring"
UNIQUE ("account_id", "bank_account_id", "transaction_recurring_id");

ALTER TABLE "funding_schedules" ADD CONSTRAINT "fk_funding_schedules_transaction_recurring"
FOREIGN KEY ("transaction_recurring_id", "account_id", "bank_account_id")
REFERENCES "transaction_recurring" ("transaction_recurring_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("transaction_recurring_id");

ALTER TABLE "funding_schedules" ADD CONSTRAINT "uq_funding_schedules_transaction_recurring"
UNIQUE ("account_id", "bank_account_id", "transaction_recurring_id");

ALTER TABLE "transaction_recurring" DROP CONSTRAINT IF EXISTS "uq_transaction_recurring_funding_schedule";
ALTER TABLE "transaction_recurring" DROP CONSTRAINT IF EXISTS "uq_transaction_recurring_spending";
ALTER TABLE "transaction_recurring" DROP CONSTRAINT IF EXISTS "ck_transaction_recurring_funding_schedule_credit";
ALTER TABLE "transaction_recurring" DROP CONSTRAINT IF EXISTS "ck_transaction_recurring_spending_debit";
ALTER TABLE "transaction_recurring" DROP CONSTRAINT IF EXISTS "fk_transaction_recurring_funding_schedule";
ALTER TABLE "transaction_recurring" DROP CONSTRAINT IF EXISTS "fk_transaction_recurring_spending";
ALTER TABLE "transaction_recurring" DROP COLUMN IF EXISTS "funding_schedule_id";
ALTER TABLE "transaction_recurring" DROP COLUMN IF EXISTS "spending_id";
