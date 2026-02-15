DELETE FROM "transaction_uploads";
DELETE FROM "files";
ALTER TABLE "files" 
ADD COLUMN "kind" TEXT NOT NULL,
DROP COLUMN "blob_uri";
