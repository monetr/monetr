ALTER TABLE "jobs" RENAME COLUMN "job_id" TO "job_id_old";
ALTER TABLE "jobs" RENAME COLUMN "job_id_new" TO "job_id";

-- Users

UPDATE "users" 
SET "login_id_new" = "logins"."login_id_new"
FROM "logins"
WHERE "logins"."login_id" = "users"."login_id";

UPDATE "users" 
SET "account_id_new" = "accounts"."account_id_new"
FROM "accounts"
WHERE "accounts"."account_id" = "users"."account_id";

-- Links

UPDATE "links"
SET "account_id_new" = "accounts"."account_id_new"
FROM "accounts"
WHERE "accounts"."account_id" = "links"."account_id";

UPDATE "links"
SET "plaid_link_id_new" = "plaid_links"."plaid_link_id_new"
FROM "plaid_links"
WHERE "plaid_links"."plaid_link_id" = "links"."plaid_link_id";

UPDATE "links"
SET "teller_link_id_new" = "teller_links"."teller_link_id_new"
FROM "teller_links"
WHERE "teller_links"."teller_link_id" = "links"."teller_link_id";

UPDATE "links"
SET "created_by_user_id_new" = "users"."user_id_new"
FROM "users"
WHERE "users"."user_id" = "links"."created_by_user_id";

-- Secrets

UPDATE "secrets"
SET "account_id_new" = "accounts"."account_id_new"
FROM "accounts"
WHERE "accounts"."account_id" = "secrets"."account_id";

-- Bank accounts

UPDATE "bank_accounts"
SET "account_id_new" = "accounts"."account_id_new"
FROM "accounts"
WHERE "accounts"."account_id" = "bank_accounts"."account_id";
