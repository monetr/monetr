ALTER TABLE "transaction_clusters" DROP CONSTRAINT "pk_transaction_clusters";
ALTER TABLE "transaction_clusters" ADD CONSTRAINT "pk_transaction_clusters" PRIMARY KEY ("transaction_cluster_id", "account_id", "bank_account_id");

-- The raw name of the centroid transaction, refreshed on every calculation.
ALTER TABLE "transaction_clusters" ADD COLUMN "original_memo" TEXT;
UPDATE "transaction_clusters" AS "txc"
SET "original_memo" = "t"."original_name"
FROM "transactions" AS "t"
WHERE "t"."transaction_id" = "txc"."centroid"
  AND "t"."account_id" = "txc"."account_id"
  AND "t"."bank_account_id" = "txc"."bank_account_id";

ALTER TABLE "transactions" ADD COLUMN "transaction_cluster_id" VARCHAR(32);

-- When a cluster is removed only the cluster ID is cleared on its transactions,
-- the account and bank account columns are part of this key too and must stay.
ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_transaction_cluster"
FOREIGN KEY ("transaction_cluster_id", "account_id", "bank_account_id")
REFERENCES "transaction_clusters" ("transaction_cluster_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("transaction_cluster_id");

-- Will make the similar transaction read very fast.
CREATE INDEX "ix_transactions_transaction_cluster"
ON "transactions" ("account_id", "bank_account_id", "transaction_cluster_id")
WHERE "transaction_cluster_id" IS NOT NULL;

UPDATE "transactions" AS "t"
SET "transaction_cluster_id" = "m"."transaction_cluster_id"
FROM (
  SELECT DISTINCT ON ("m"."transaction_id", "txc"."account_id", "txc"."bank_account_id")
    "m"."transaction_id",
    "txc"."account_id",
    "txc"."bank_account_id",
    "txc"."transaction_cluster_id"
  FROM "transaction_clusters" AS "txc"
  CROSS JOIN LATERAL UNNEST("txc"."members") AS "m"("transaction_id")
  ORDER BY "m"."transaction_id", "txc"."account_id", "txc"."bank_account_id", "txc"."transaction_cluster_id"
) AS "m"
WHERE "t"."transaction_id" = "m"."transaction_id"
  AND "t"."account_id" = "m"."account_id"
  AND "t"."bank_account_id" = "m"."bank_account_id";
