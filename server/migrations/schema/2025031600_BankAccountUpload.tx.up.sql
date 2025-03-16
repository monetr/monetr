ALTER TABLE "bank_accounts"
ADD COLUMN "upload_identifier" VARCHAR(200),
ADD CONSTRAINT "uq_bank_accounts_upload_identifier" UNIQUE ("account_id", "link_id", "upload_identifier");
