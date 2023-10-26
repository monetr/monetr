ALTER TABLE "transactions" ADD COLUMN "currency" TEXT DEFAULT 'USD';
UPDATE "transactions" SET "currency"='USD' WHERE "currency" IS NULL;
ALTER TABLE "transactions" ALTER COLUMN "currency" SET NOT NULL;

