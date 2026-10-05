ALTER TABLE "funding_schedules" DROP CONSTRAINT IF EXISTS "uq_funding_schedules_transaction_recurring";
ALTER TABLE "funding_schedules" DROP CONSTRAINT IF EXISTS "fk_funding_schedules_transaction_recurring";
ALTER TABLE "funding_schedules" DROP COLUMN IF EXISTS "transaction_recurring_id";
