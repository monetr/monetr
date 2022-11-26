DROP INDEX IF EXISTS "ix_transactions_soft_delete";
ALTER TABLE "transactions" DROP COLUMN "deleted_at";
