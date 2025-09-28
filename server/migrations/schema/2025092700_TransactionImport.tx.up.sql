CREATE TABLE "transaction_imports" (
  "transaction_import_id" VARCHAR(32) NOT NULL,
  "account_id"            VARCHAR(32) NOT NULL,
  "link_id"               VARCHAR(32) NOT NULL,
  "file_id"               VARCHAR(32) NOT NULL,
  "status"                VARCHAR(32) NOT NULL,
  "expires_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_by"            VARCHAR(32) NOT NULL,
  "updated_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT "pk_transaction_imports" PRIMARY KEY ("transaction_import_id", "account_id"),
  CONSTRAINT "fk_transaction_imports_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_transaction_imports_link" FOREIGN KEY ("link_id", "account_id") REFERENCES "links" ("link_id", "account_id"),
  CONSTRAINT "fk_transaction_imports_file" FOREIGN KEY ("file_id", "account_id") REFERENCES "files" ("file_id", "account_id"),
  CONSTRAINT "fk_transaction_imports_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id")
);

CREATE TABLE "transaction_import_items" (
  "transaction_import_item_id" VARCHAR(32) NOT NULL,
  "account_id"                 VARCHAR(32) NOT NULL,
  "transaction_import_id"      VARCHAR(32) NOT NULL,
  "bank_account_id"            VARCHAR(32),
  "name"                       TEXT NOT NULL,
  "currency"                   TEXT NOT NULL,
  "include"                    BOOLEAN NOT NULL,
  "created_at"                 TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT "pk_transaction_import_items" PRIMARY KEY ("transaction_import_item_id", "account_id", "transaction_import_id"),
  CONSTRAINT "fk_transaction_import_items_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_transaction_import_items_import" FOREIGN KEY ("transaction_import_id", "account_id") REFERENCES "transaction_imports" ("transaction_import_id", "account_id"),
  CONSTRAINT "fk_transaction_import_items_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id")
);
