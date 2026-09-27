ALTER TABLE "accounts"
    ADD COLUMN "subscription_status" TEXT;

-- For accounts that do have an active subscription, make sure the status is "active".
UPDATE "accounts"
SET "subscription_status" = 'active'
WHERE "subscription_active_until" > CURRENT_TIMESTAMP;
-- For accounts that do not currently have an active subscription, just update the status to be "canceled".
UPDATE "accounts"
SET "subscription_status" = 'canceled'
WHERE "subscription_active_until" <= CURRENT_TIMESTAMP;

