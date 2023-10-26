ALTER TABLE "accounts" ADD COLUMN "trial_ends_at" TIMESTAMPTZ;

ALTER TABLE "accounts" ADD COLUMN "created_at" TIMESTAMPTZ;
-- Since we don't know when the account was actually created, we are going to just assume it was created now. And then
-- we can just have correct data going forward.
UPDATE "accounts" SET "created_at" = now() WHERE "created_at" IS NULL;
ALTER TABLE "accounts" ALTER COLUMN "created_at" SET NOT NULL;

