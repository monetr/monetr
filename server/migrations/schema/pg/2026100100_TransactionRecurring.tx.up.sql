CREATE TABLE "transaction_recurring" (
  "transaction_recurring_id" VARCHAR(32) NOT NULL,
  "account_id"               VARCHAR(32) NOT NULL,
  "bank_account_id"          VARCHAR(32) NOT NULL,
  "transaction_cluster_id"   VARCHAR(32) NOT NULL,
  "direction"                TEXT        NOT NULL,
  "window_type"              TEXT        NOT NULL,
  "ruleset"                  TEXT        NOT NULL,
  "first"                    TIMESTAMPTZ NOT NULL,
  "last"                     TIMESTAMPTZ NOT NULL,
  "next"                     TIMESTAMPTZ NOT NULL,
  "ended"                    BOOLEAN     NOT NULL DEFAULT FALSE,
  "confidence"               REAL        NOT NULL,
  "amounts"                  JSONB       NOT NULL,
  "last_amount"              BIGINT      NOT NULL,
  "created_at"               TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now(),
  "updated_at"               TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now(),
  CONSTRAINT "pk_transaction_recurring"                     PRIMARY KEY ("transaction_recurring_id", "account_id", "bank_account_id"),
  CONSTRAINT "uq_transaction_recurring_cluster_direction"   UNIQUE ("account_id", "bank_account_id", "transaction_cluster_id", "direction"),
  CONSTRAINT "fk_transaction_recurring_account"             FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_recurring_bank_account"        FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_recurring_transaction_cluster" FOREIGN KEY ("transaction_cluster_id", "account_id", "bank_account_id") REFERENCES "transaction_clusters" ("transaction_cluster_id", "account_id", "bank_account_id") ON DELETE CASCADE
);

CREATE INDEX "ix_transaction_recurring_transaction_cluster"
ON "transaction_recurring" ("account_id", "bank_account_id", "transaction_cluster_id");

ALTER TABLE "transactions" ADD COLUMN "transaction_recurring_id" VARCHAR(32);

-- Only null out the recurring ID when a recurring transaction is deleted,
-- account and bank account are part of this key too and need to stay
ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_transaction_recurring"
FOREIGN KEY ("transaction_recurring_id", "account_id", "bank_account_id")
REFERENCES "transaction_recurring" ("transaction_recurring_id", "account_id", "bank_account_id")
ON DELETE SET NULL ("transaction_recurring_id");

CREATE INDEX "ix_transactions_transaction_recurring"
ON "transactions" ("account_id", "bank_account_id", "transaction_recurring_id")
WHERE "transaction_recurring_id" IS NOT NULL;
