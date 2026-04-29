CREATE TABLE "transaction_import_mappings" (
  "transaction_import_mapping_id" VARCHAR(32) NOT NULL,
  "account_id"                    VARCHAR(32) NOT NULL,
  "signature"                     TEXT        NOT NULL, -- No unique enforcement!
  "mapping"                       JSONB       NOT NULL,
  "created_by"                    VARCHAR(32) NOT NULL,
  "created_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "updated_at"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  CONSTRAINT "pk_transaction_import_mappings"            PRIMARY KEY ("transaction_import_mapping_id", "account_id"),
  CONSTRAINT "fk_transaction_import_mappings_account"    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_transaction_import_mappings_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id") ON DELETE CASCADE
);
