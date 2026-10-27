CREATE TABLE "transaction_import_previews" (
  "transaction_import_preview_id" VARCHAR(32) NOT NULL,
  "account_id"                    VARCHAR(32) NOT NULL,
  "bank_account_id"               VARCHAR(32) NOT NULL,
  "transaction_import_id"         VARCHAR(32) NOT NULL,
  "rows"                          JSONB NOT NULL,
  "available_balance"             BIGINT NOT NULL,
  "current_balance"               BIGINT NOT NULL,
  "created_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "updated_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  CONSTRAINT "pk_transaction_import_previews"              PRIMARY KEY ("transaction_import_preview_id", "account_id", "bank_account_id"),
  CONSTRAINT "uq_transaction_import_previews"              UNIQUE ("account_id", "bank_account_id", "transaction_import_id"),
  CONSTRAINT "fk_transaction_import_previews_account"      FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_import_previews_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_import_previews_import"       FOREIGN KEY ("transaction_import_id", "account_id", "bank_account_id") REFERENCES "transaction_imports" ("transaction_import_id", "account_id", "bank_account_id") ON DELETE CASCADE
);
