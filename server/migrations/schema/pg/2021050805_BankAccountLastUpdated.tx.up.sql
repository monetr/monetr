ALTER TABLE "bank_accounts" ADD COLUMN "last_updated" TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'UTC');
UPDATE "bank_accounts" SET "last_updated" = (NOW() AT TIME ZONE 'UTC');
ALTER TABLE "bank_accounts" ALTER COLUMN "last_updated" SET NOT NULL;