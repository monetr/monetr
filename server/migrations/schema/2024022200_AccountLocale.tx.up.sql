ALTER TABLE "accounts" ADD COLUMN "locale" TEXT;
UPDATE "accounts" SET "locale" = 'en_US';
ALTER TABLE "accounts" ALTER COLUMN "locale" SET NOT NULL;
