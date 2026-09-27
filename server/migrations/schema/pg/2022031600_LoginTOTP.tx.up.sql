-- Add the totp string column to the logins table, make it optional though. It will only be populated for those with it
-- enabled.
ALTER TABLE "logins" ADD COLUMN "totp" TEXT;
ALTER TABLE "logins" ADD COLUMN "totp_enabled_at" TIMESTAMPTZ;
