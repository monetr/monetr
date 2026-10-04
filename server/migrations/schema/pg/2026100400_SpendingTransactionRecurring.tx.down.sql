ALTER TABLE "spending" DROP CONSTRAINT IF EXISTS "uq_spending_transaction_recurring";
ALTER TABLE "spending" DROP CONSTRAINT IF EXISTS "fk_spending_transaction_recurring";
ALTER TABLE "spending" DROP COLUMN IF EXISTS "transaction_recurring_id";
