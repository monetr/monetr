-- This column is there primarily for debugging and is not meant to be used for
-- foreign keys or anything yet.
ALTER TABLE "transaction_clusters" ADD COLUMN "centroid" VARCHAR(30);
