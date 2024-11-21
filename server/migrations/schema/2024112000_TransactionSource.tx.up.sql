ALTER TABLE "transactions" ADD COLUMN "source" TEXT;

UPDATE "transactions" SET "source" = 'plaid' 
WHERE "plaid_transaction_id" IS NOT NULL OR "pending_plaid_transaction_id" IS NOT NULL;

UPDATE "transactions" SET "source" = 'upload'
WHERE "source" IS NULL AND "upload_identifier" IS NOT NULL;

UPDATE "transactions" SET "source" = 'manual'
WHERE "source" IS NULL;

ALTER TABLE "transactions" ALTER COLUMN "source" SET NOT NULL;
