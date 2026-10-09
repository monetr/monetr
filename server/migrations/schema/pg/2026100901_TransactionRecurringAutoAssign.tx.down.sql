ALTER TABLE "transaction_recurring" DROP CONSTRAINT IF EXISTS "ck_transaction_recurring_auto_assign_spending";
ALTER TABLE "transaction_recurring" DROP COLUMN IF EXISTS "auto_assign";
