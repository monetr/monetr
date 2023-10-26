ALTER TABLE "transactions" ADD COLUMN "deleted_at" TIMESTAMP WITH TIME ZONE NULL;

CREATE INDEX "ix_transactions_soft_delete"
ON "transactions" ("account_id", "bank_account_id", "date" DESC, "transaction_id" DESC)
WHERE "deleted_at" IS NULL;
