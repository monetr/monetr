DROP INDEX IF EXISTS "ix_transactions_transaction_cluster";
ALTER TABLE "transactions" DROP CONSTRAINT IF EXISTS "fk_transactions_transaction_cluster";
ALTER TABLE "transactions" DROP COLUMN IF EXISTS "transaction_cluster_id";

ALTER TABLE "transaction_clusters" DROP CONSTRAINT "pk_transaction_clusters";
ALTER TABLE "transaction_clusters" ADD CONSTRAINT "pk_transaction_clusters" PRIMARY KEY ("transaction_cluster_id", "account_id");
