ALTER TABLE "transaction_clusters" ADD COLUMN "original_name" TEXT, ADD COLUMN "updated_at" TIMESTAMP WITHOUT TIME ZONE;
UPDATE "transaction_clusters" SET "original_name" = "name", "updated_at" = "created_at";
ALTER TABLE "transaction_clusters" ALTER COLUMN "original_name" SET NOT NULL;
ALTER TABLE "transaction_clusters" ALTER COLUMN "updated_at" SET NOT NULL;
ALTER TABLE "transaction_clusters" ADD CONSTRAINT "uq_transaction_cluster_persist" UNIQUE ("account_id", "bank_account_id", "signature", "centroid");
