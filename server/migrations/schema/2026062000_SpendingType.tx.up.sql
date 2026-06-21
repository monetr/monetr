ALTER TABLE "spending" DROP CONSTRAINT "uq_spending_type_name";
ALTER TABLE "spending" ADD COLUMN "spending_type_new" TEXT;
UPDATE "spending" SET "spending_type_new" = 'expense' WHERE "spending_type" = 0;
UPDATE "spending" SET "spending_type_new" = 'goal' WHERE "spending_type" = 1;
-- Everything else?
UPDATE "spending" SET "spending_type_new" = 'goal' WHERE "spending_type_new" IS NULL;
ALTER TABLE "spending" ALTER COLUMN "spending_type_new" SET NOT NULL;
ALTER TABLE "spending" DROP COLUMN "spending_type";
ALTER TABLE "spending" RENAME COLUMN "spending_type_new" TO "spending_type";
ALTER TABLE "spending" ADD CONSTRAINT "uq_spending_type_name" UNIQUE ("bank_account_id", "spending_type", "name");
