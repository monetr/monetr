ALTER TABLE "plaid_transactions" ALTER COLUMN "name" SET DATA TYPE TEXT;
ALTER TABLE "plaid_transactions" ALTER COLUMN "merchant_name" SET DATA TYPE TEXT;
ALTER TABLE "transactions" ALTER COLUMN "name" SET DATA TYPE TEXT;
ALTER TABLE "transactions" ALTER COLUMN "original_name" SET DATA TYPE TEXT;
ALTER TABLE "transactions" ALTER COLUMN "merchant_name" SET DATA TYPE TEXT;
ALTER TABLE "transactions" ALTER COLUMN "original_merchant_name" SET DATA TYPE TEXT;
