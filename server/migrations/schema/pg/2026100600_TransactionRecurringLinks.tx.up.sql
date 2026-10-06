-- Flip the link between recurring transactions and the spending or funding
-- schedule tracking them. The recurring transaction now points at the spending
-- object (debits) or the funding schedule (credits) instead of the other way
-- around.
ALTER TABLE "transaction_recurring" ADD COLUMN "spending_id" VARCHAR(32);
ALTER TABLE "transaction_recurring" ADD COLUMN "funding_schedule_id" VARCHAR(32);

-- Only null out the spending or funding schedule ID when the thing it points at
-- is deleted, account and bank account are part of this key too and need to
-- stay.
ALTER TABLE "transaction_recurring" ADD CONSTRAINT "fk_transaction_recurring_spending"
FOREIGN KEY ("spending_id", "account_id", "bank_account_id")
REFERENCES "spending" ("spending_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("spending_id");

ALTER TABLE "transaction_recurring" ADD CONSTRAINT "fk_transaction_recurring_funding_schedule"
FOREIGN KEY ("funding_schedule_id", "account_id", "bank_account_id")
REFERENCES "funding_schedules" ("funding_schedule_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("funding_schedule_id");

-- Carry over any existing links before the old columns are dropped.
UPDATE "transaction_recurring" AS "recurring"
SET "spending_id" = "spending"."spending_id"
FROM "spending"
WHERE "spending"."account_id" = "recurring"."account_id"
  AND "spending"."bank_account_id" = "recurring"."bank_account_id"
  AND "spending"."transaction_recurring_id" = "recurring"."transaction_recurring_id";

UPDATE "transaction_recurring" AS "recurring"
SET "funding_schedule_id" = "funding_schedules"."funding_schedule_id"
FROM "funding_schedules"
WHERE "funding_schedules"."account_id" = "recurring"."account_id"
  AND "funding_schedules"."bank_account_id" = "recurring"."bank_account_id"
  AND "funding_schedules"."transaction_recurring_id" = "recurring"."transaction_recurring_id";

-- Spending only tracks money going out and funding schedules only track money
-- coming in. These are added after the backfill on purpose so that bad data
-- fails the migration instead of being silently dropped.
ALTER TABLE "transaction_recurring" ADD CONSTRAINT "ck_transaction_recurring_spending_debit"
CHECK ("spending_id" IS NULL OR "direction" = 'debit');

ALTER TABLE "transaction_recurring" ADD CONSTRAINT "ck_transaction_recurring_funding_schedule_credit"
CHECK ("funding_schedule_id" IS NULL OR "direction" = 'credit');

-- For now a spending object or funding schedule can only be tracked by a single
-- recurring transaction. Nulls are distinct so unlinked recurring transactions
-- are fine.
ALTER TABLE "transaction_recurring" ADD CONSTRAINT "uq_transaction_recurring_spending"
UNIQUE ("account_id", "bank_account_id", "spending_id");

ALTER TABLE "transaction_recurring" ADD CONSTRAINT "uq_transaction_recurring_funding_schedule"
UNIQUE ("account_id", "bank_account_id", "funding_schedule_id");

ALTER TABLE "spending" DROP CONSTRAINT "uq_spending_transaction_recurring";
ALTER TABLE "spending" DROP CONSTRAINT "fk_spending_transaction_recurring";
ALTER TABLE "spending" DROP COLUMN "transaction_recurring_id";

ALTER TABLE "funding_schedules" DROP CONSTRAINT "uq_funding_schedules_transaction_recurring";
ALTER TABLE "funding_schedules" DROP CONSTRAINT "fk_funding_schedules_transaction_recurring";
ALTER TABLE "funding_schedules" DROP COLUMN "transaction_recurring_id";
