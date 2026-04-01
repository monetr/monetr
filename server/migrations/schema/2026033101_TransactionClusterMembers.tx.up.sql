ALTER TABLE "transaction_clusters" DROP CONSTRAINT "pk_transaction_clusters";
ALTER TABLE "transaction_clusters" ADD CONSTRAINT "pk_transaction_clusters" PRIMARY KEY ("transaction_cluster_id", "account_id", "bank_account_id");

-- TODO Add down migration or make all of this conditional
CREATE TABLE "transaction_cluster_members" (
  "transaction_id"                VARCHAR(32) NOT NULL,
  "account_id"                    VARCHAR(32) NOT NULL,
  "bank_account_id"               VARCHAR(32) NOT NULL,
  "transaction_cluster_id"        VARCHAR(32) NOT NULL,
  "updated_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  "created_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  CONSTRAINT "pk_transaction_cluster_members" PRIMARY KEY ("transaction_id", "account_id", "bank_account_id"),
  CONSTRAINT "fk_transaction_cluster_members_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_cluster_members_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_cluster_members_cluster" FOREIGN KEY ("transaction_cluster_id", "account_id", "bank_account_id") REFERENCES "transaction_clusters" ("transaction_cluster_id", "account_id", "bank_account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_cluster_members_transaction" FOREIGN KEY ("transaction_id", "account_id", "bank_account_id") REFERENCES "transactions" ("transaction_id", "account_id", "bank_account_id") ON DELETE CASCADE
);

-- Will make the similar transaction read very fast.
CREATE INDEX "ix_transaction_cluster_members_bank_account"
ON "transaction_cluster_members" ("account_id", "bank_account_id", "transaction_cluster_id");

INSERT INTO "transaction_cluster_members" ("transaction_id", "account_id", "bank_account_id", "transaction_cluster_id")
SELECT DISTINCT ON ("m"."transaction_id", "txc"."account_id", "txc"."bank_account_id")
  "m"."transaction_id",
  "txc"."account_id",
  "txc"."bank_account_id",
  "txc"."transaction_cluster_id"
FROM "transaction_clusters" AS "txc"
CROSS JOIN LATERAL UNNEST("txc"."members") AS "m"("transaction_id")
INNER JOIN "transactions" AS "t"
  ON "t"."transaction_id" = "m"."transaction_id"
  AND "t"."account_id" = "txc"."account_id"
  AND "t"."bank_account_id" = "txc"."bank_account_id"
ORDER BY "m"."transaction_id", "txc"."account_id", "txc"."bank_account_id", "txc"."transaction_cluster_id";

