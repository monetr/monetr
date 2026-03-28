ALTER TABLE "transaction_clusters" DROP CONSTRAINT "pk_transaction_clusters";
ALTER TABLE "transaction_clusters" ADD CONSTRAINT "pk_transaction_clusters" PRIMARY KEY ("transaction_cluster_id", "account_id", "bank_account_id");

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

INSERT INTO "transaction_cluster_members" ("transaction_id", "account_id", "bank_account_id", "transaction_cluster_id")
SELECT
  array_agg("txc"."members") as "transaction_id",
  "txc"."account_id",
  "txc"."bank_account_id",
  "txc"."transaction_cluster_id"
FROM "transaction_clusters" AS "txc"
GROUP BY
  "txc"."account_id",
  "txc"."bank_account_id",
  "txc"."transaction_cluster_id";
