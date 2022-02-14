ALTER TABLE "products" DROP COLUMN "features";
ALTER TABLE "products" ADD COLUMN "features" BIGINT;
ALTER TABLE "products" ALTER COLUMN "features" SET NOT NULL;
