DROP INDEX IF EXISTS "ix_transactions_transaction_recurring";
ALTER TABLE "transactions" DROP CONSTRAINT IF EXISTS "fk_transactions_transaction_recurring";
ALTER TABLE "transactions" DROP COLUMN IF EXISTS "transaction_recurring_id";
DROP TABLE IF EXISTS "transaction_recurring";
