ALTER TABLE "spending" ADD COLUMN "spending_type_new" TEXT NOT NULL;
UPDATE "spending" SET "spending_type_new" = 'expense' WHERE "spending_type" = 0;
UPDATE "spending" SET "spending_type_new" = 'goal' WHERE "spending_type" = 1;
ALTER TABLE "spending" DROP CONSTRAINT "uq_spending_type_name";
ALTER TABLE "spending" DROP COLUMN "spending_type";
ALTER TABLE "spending" RENAME COLUMN "spending_type_new" TO "spending_type";
-- Need to re-add this since we are changing the old column!
ALTER TABLE "spending" ADD CONSTRAINT "uq_spending_type_name" UNIQUE ("bank_account_id", "spending_type", "name");

