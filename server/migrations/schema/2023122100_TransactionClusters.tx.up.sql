CREATE TABLE "transaction_clusters" (
  "transaction_cluster_id" TEXT                     NOT NULL,
  "account_id"             BIGINT                   NOT NULL,
  "bank_account_id"        BIGINT                   NOT NULL,
  "name"                   TEXT                     NOT NULL,
  "members"                BIGINT[]                 NOT NULL,
  "created_at"             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  CONSTRAINT "pk_transaction_clusters"              PRIMARY KEY ("transaction_cluster_id"),
  CONSTRAINT "fk_transaction_clusters_account"      FOREIGN KEY ("account_id")                    REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_transaction_clusters_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id")
);

-- For querying by the members contents.
CREATE INDEX "ix_transaction_clusters_members" ON "transaction_clusters" USING GIN ("members");
-- For narrowing down the results to a single bank account.
CREATE INDEX "ix_transaction_clusters_bank_account" ON "transaction_clusters" ("account_id", "bank_account_id");
