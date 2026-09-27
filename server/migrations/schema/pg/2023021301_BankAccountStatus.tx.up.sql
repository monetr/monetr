ALTER TABLE "bank_accounts" ADD COLUMN "status" TEXT DEFAULT 'active';
UPDATE "bank_accounts" SET "status" = 'active';
ALTER TABLE "bank_accounts" ALTER COLUMN "status" SET NOT NULL;
