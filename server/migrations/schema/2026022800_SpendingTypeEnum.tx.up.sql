ALTER TABLE "spending" ADD COLUMN "spending_type_new" TEXT NOT NULL DEFAULT 'expense';
UPDATE "spending" SET "spending_type_new" = 'expense' WHERE "spending_type" = 0;
UPDATE "spending" SET "spending_type_new" = 'goal' WHERE "spending_type" = 1;
UPDATE "spending" SET "spending_type_new" = 'overflow' WHERE "spending_type" = 2;
ALTER TABLE "spending" DROP COLUMN "spending_type";
ALTER TABLE "spending" RENAME COLUMN "spending_type_new" TO "spending_type";
