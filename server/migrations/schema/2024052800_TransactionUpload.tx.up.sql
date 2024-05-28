ALTER TABLE "transactions" 
ADD COLUMN "upload_identifier" VARCHAR(200),
ADD CONSTRAINT "uq_transactions_upload_identifier" UNIQUE ("account_id", "bank_account_id", "upload_identifier");
