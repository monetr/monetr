ALTER TABLE "transactions" 
ADD COLUMN "upload_identifier" VARCHAR(200),
ADD CONSTRAINT "uq_transactions_upload_identifier" UNIQUE ("account_id", "bank_account_id", "upload_identifier");

CREATE TABLE "transaction_uploads" (
  "transaction_upload_id" VARCHAR(32) NOT NULL,
  "account_id"            VARCHAR(32) NOT NULL,
  "bank_account_id"       VARCHAR(32) NOT NULL,
  "file_id"               VARCHAR(32) NOT NULL,
  "status"                VARCHAR(16) NOT NULL DEFAULT 'pending',
  "error"                 TEXT,
  "created_at"            TIMESTAMP WITH TIME ZONE NOT NULL,
  "created_by"            VARCHAR(32) NOT NULL,
  "processed_at"          TIMESTAMP WITH TIME ZONE,
  "completed_at"          TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_transaction_uploads" PRIMARY KEY ("transaction_upload_id", "bank_account_id", "account_id"),
  CONSTRAINT "fk_transaction_uploads_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_transaction_uploads_bank_account" FOREIGN KEY ("bank_account_id", "account_id") REFERENCES "bank_accounts" ("bank_account_id", "account_id"),
  CONSTRAINT "fk_transaction_uploads_file" FOREIGN KEY ("file_id", "account_id") REFERENCES "files" ("file_id", "account_id"),
  CONSTRAINT "fk_transaction_uploads_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id")
);
