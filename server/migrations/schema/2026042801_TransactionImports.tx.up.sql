CREATE TABLE "transaction_imports" (
  "transaction_import_id"         VARCHAR(32) NOT NULL,
  "account_id"                    VARCHAR(32) NOT NULL,
  "bank_account_id"               VARCHAR(32) NOT NULL,
  "file_id"                       VARCHAR(32) NOT NULL,
  "transaction_import_mapping_id" VARCHAR(32),
  "headers"                       TEXT[]      NOT NULL,
  "status"                        TEXT        NOT NULL,
  "created_by"                    VARCHAR(32) NOT NULL,
  "created_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "updated_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "completed_at"                  TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_transaction_imports"              PRIMARY KEY ("transaction_import_id", "account_id", "bank_account_id"),
  CONSTRAINT "fk_transaction_imports_account"      FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_imports_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_imports_file"         FOREIGN KEY ("file_id", "account_id") REFERENCES "files" ("file_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_imports_mapping"      FOREIGN KEY ("transaction_import_mapping_id", "account_id") REFERENCES "transaction_import_mappings" ("transaction_import_mapping_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_imports_created_by"   FOREIGN KEY ("created_by") REFERENCES "users" ("user_id") ON DELETE CASCADE
);
