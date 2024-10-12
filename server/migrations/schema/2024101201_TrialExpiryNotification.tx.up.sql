-- Add the new notification column.
ALTER TABLE "accounts"
ADD COLUMN "trial_expiry_notification_sent_at" TIMESTAMP WITH TIME ZONE;

-- But make sure that we don't send the notification for anyone who has already
-- ended their trial.
UPDATE "accounts"
SET "trial_expiry_notification_sent_at" = "trial_ends_at"
WHERE "trial_ends_at" <= now();


-- I need to be able to know which user is the owner of an account, moentr will
-- someday support multiple users per account and will need to know who is the
-- owner. I could add this as a field to the accounts table like `owner_id` but
-- then you end up with a circular reference between the two tables. Very evil.
-- Instead, adding the role to the user table, then imposing a unique
-- constraint on the account IDs of the user table but only for the rows who's
-- role is `owner` allows us to make sure that there is only ever a single
-- owner. Note however, that it is possible in this world to have _no owners_.
-- Also very evil. There is no winning.
CREATE TYPE user_role AS ENUM ('owner', 'member');

-- Add our role column and remove a confusing billing column. User is not the
-- source of truth for cutomer ID.
ALTER TABLE "users" 
ADD COLUMN "role" user_role, 
DROP COLUMN "stripe_customer_id";

-- Backfill the owners based on the users who are the oldest (lowest ID) per
-- account.
UPDATE "users"
SET "role" = 'owner'
WHERE "user_id" IN (
  -- Get the user with the lowest ID per account, this is the user that was
  -- created first and should be considered the "owner".
  SELECT min("user_id")
  FROM "users"
  GROUP BY "account_id"
);

-- Backfill members (even though there shouldn't be any) by whoever does not
-- still have a role.
UPDATE "users"
SET "role" = 'member'
WHERE "role" IS NULL;

-- Then make sure that we will always have a role going forward.
ALTER TABLE "users" ALTER COLUMN "role" SET NOT NULL;

-- And that we cannot have more than a single owner per account.
CREATE UNIQUE INDEX "ix_uq_users_role_owner" ON "users" ("account_id", "role") 
WHERE ("role" = 'owner');

