ALTER TABLE "plaid_syncs"
ADD COLUMN "account_id" BIGINT;

UPDATE "plaid_syncs"
SET "account_id" = "links"."account_id"
FROM "plaid_links"
INNER JOIN "links" ON "links"."plaid_link_id" = "plaid_links"."plaid_link_id"
WHERE "plaid_links"."plaid_link_id" = "plaid_syncs"."plaid_link_id";

-- Add and backfill the account ID column to the Plaid links table.
ALTER TABLE "plaid_links" 
ADD COLUMN "account_id" BIGINT,
ADD COLUMN "status" INT,
ADD COLUMN "error_code" TEXT,
ADD COLUMN "expiration_date" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "new_accounts_available" BOOLEAN,
ADD COLUMN "last_manual_sync" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "last_successful_update" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "last_attempted_update" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "updated_at" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "created_at" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "created_by_user_id" BIGINT,
DROP COLUMN "use_plaid_sync";

UPDATE "plaid_links" 
SET "account_id"             = "links"."account_id",
    "institution_id"         = "links"."plaid_institution_id",
    "institution_name"       = "links"."institution_name",
    "status"                 = "links"."link_status",
    "error_code"             = "links"."error_code",
    "expiration_date"        = "links"."expiration_date",
    "new_accounts_available" = "links"."plaid_new_accounts_available",
    "last_manual_sync"       = "links"."last_manual_sync",
    "last_successful_update" = "links"."last_successful_update",
    "last_attempted_update"  = "links"."last_attempted_update",
    "updated_at"             = "links"."updated_at",
    "created_at"             = "links"."created_at",
    "created_by_user_id"     = "links"."created_by_user_id"
FROM "links" WHERE "links"."plaid_link_id" = "plaid_links"."plaid_link_id";

ALTER TABLE "plaid_links" 
ALTER COLUMN "account_id" SET NOT NULL,
ALTER COLUMN "institution_id" SET NOT NULL,
ALTER COLUMN "institution_name" SET NOT NULL,
ALTER COLUMN "status" SET NOT NULL,
ALTER COLUMN "status" SET DEFAULT 0,
ALTER COLUMN "new_accounts_available" SET NOT NULL,
ALTER COLUMN "new_accounts_available" SET DEFAULT false,
ALTER COLUMN "updated_at" SET NOT NULL,
ALTER COLUMN "updated_at" SET DEFAULT now(),
ALTER COLUMN "created_at" SET NOT NULL,
ALTER COLUMN "created_at" SET DEFAULT now(),
ALTER COLUMN "created_by_user_id" SET NOT NULL,
DROP CONSTRAINT "pk_plaid_links" CASCADE, 
ADD PRIMARY KEY ("plaid_link_id", "account_id"),
ADD CONSTRAINT "fk_plaid_links_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id");

ALTER TABLE "plaid_syncs"
DROP CONSTRAINT "pk_plaid_syncs" CASCADE, 
ADD PRIMARY KEY ("plaid_sync_id", "account_id"),
ADD CONSTRAINT "fk_plaid_syncs_plaid_link" FOREIGN KEY ("plaid_link_id", "account_id") REFERENCES "plaid_links" ("plaid_link_id", "account_id"),
ADD CONSTRAINT "fk_plaid_syncs_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id");

ALTER TABLE "links"
DROP COLUMN "link_status",
DROP COLUMN "last_successful_update",
DROP COLUMN "expiration_date",
DROP COLUMN "plaid_institution_id",
DROP COLUMN "last_manual_sync",
DROP COLUMN "last_attempted_update",
DROP COLUMN "plaid_new_accounts_available",
DROP COLUMN "error_code",
DROP COLUMN "custom_institution_name",
ADD CONSTRAINT "fk_links_plaid_link" FOREIGN KEY ("plaid_link_id", "account_id") REFERENCES "plaid_links" ("plaid_link_id", "account_id");

-- Now create the new plaid bank accounts table and backfill it.
CREATE TABLE "plaid_bank_accounts" (
  "plaid_bank_account_id" BIGSERIAL NOT NULL,
  "account_id"            BIGINT    NOT NULL,
  "plaid_link_id"         BIGINT    NOT NULL,
  "plaid_id"              TEXT      NOT NULL,
  "name"                  TEXT      NOT NULL,
  "official_name"         TEXT,
  "mask"                  TEXT,
  "available_balance"     BIGINT    NOT NULL,
  "current_balance"       BIGINT    NOT NULL,
  "limit_balance"         BIGINT,
  "created_at"            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  "created_by_user_id"    BIGINT NOT NULL,

  CONSTRAINT "pk_plaid_bank_accounts"                          PRIMARY KEY ("plaid_bank_account_id", "account_id"),
  CONSTRAINT "fk_plaid_bank_accounts_account"                  FOREIGN KEY ("account_id")                  REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_plaid_bank_accounts_plaid_link"               FOREIGN KEY ("plaid_link_id", "account_id") REFERENCES "plaid_links" ("plaid_link_id", "account_id"),
  CONSTRAINT "fk_plaid_bank_accounts_users_created_by_user_id" FOREIGN KEY ("created_by_user_id")          REFERENCES "users" ("user_id")
);

INSERT INTO "plaid_bank_accounts" 
  ("account_id", "plaid_link_id", "plaid_id", "name", "official_name", "mask", "available_balance", "current_balance", "created_by_user_id")
SELECT 
  "bank_accounts"."account_id",
  "links"."plaid_link_id",
  "bank_accounts"."plaid_account_id" AS "plaid_id",
  "bank_accounts"."plaid_name" AS "name",
  "bank_accounts"."plaid_official_name" AS "official_name",
  "bank_accounts"."mask",
  "bank_accounts"."available_balance",
  "bank_accounts"."current_balance",
  "links"."created_by_user_id"
FROM "bank_accounts"
INNER JOIN "links" ON "links"."link_id" = "bank_accounts"."link_id" AND 
                      "links"."account_id" = "bank_accounts"."account_id"
WHERE "bank_accounts"."plaid_account_id" IS NOT NULL AND
      "links"."plaid_link_id" IS NOT NULL;

-- Then update the bank accounts table with the new stuff.
ALTER TABLE "bank_accounts" 
ADD COLUMN "original_name" TEXT,
ADD COLUMN "plaid_bank_account_id" BIGINT,
ADD COLUMN "updated_at" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "created_at" TIMESTAMP WITH TIME ZONE,
ADD COLUMN "created_by_user_id" BIGINT;

UPDATE "bank_accounts" 
SET "plaid_bank_account_id" = "plaid_bank_accounts"."plaid_bank_account_id"
FROM "plaid_bank_accounts"
WHERE "plaid_bank_accounts"."plaid_id" = "bank_accounts"."plaid_account_id" AND
      "plaid_bank_accounts"."account_id" = "bank_accounts"."account_id";

UPDATE "bank_accounts" 
SET "original_name" = "plaid_name",
    "updated_at" = "links"."updated_at",
    "created_at" = "links"."created_at",
    "created_by_user_id" = "links"."created_by_user_id"
FROM "links"
WHERE "links"."link_id" = "bank_accounts"."link_id" AND
      "links"."account_id" = "bank_accounts"."account_id";

-- Then cleanup the columns that are no longer meant to be read from this table.
ALTER TABLE "bank_accounts"
DROP COLUMN "plaid_account_id",
DROP COLUMN "plaid_name",
DROP COLUMN "plaid_official_name",
ALTER COLUMN "updated_at" SET NOT NULL,
ALTER COLUMN "updated_at" SET DEFAULT now(),
ALTER COLUMN "created_at" SET NOT NULL,
ALTER COLUMN "created_at" SET DEFAULT now(),
ALTER COLUMN "created_by_user_id" SET NOT NULL,
ADD CONSTRAINT "fk_bank_accounts_plaid_bank_accounts" FOREIGN KEY ("plaid_bank_account_id", "account_id") REFERENCES "plaid_bank_accounts" ("plaid_bank_account_id", "account_id");


CREATE TABLE "plaid_transactions" (
  "plaid_transaction_id"         BIGSERIAL NOT NULL,
  "account_id"                   BIGINT    NOT NULL,
  "plaid_bank_account_id"        BIGINT    NOT NULL,
  "plaid_id"                     TEXT      NOT NULL,
  "pending_plaid_id"             TEXT,
  "categories"                   TEXT[],
  "date"                         TIMESTAMP WITH TIME ZONE NOT NULL,
  "authorized_date"              TIMESTAMP WITH TIME ZONE,
  "name"                         TEXT    NOT NULL,
  "merchant_name"                TEXT,
  "amount"                       BIGINT  NOT NULL,
  "currency"                     TEXT    NOT NULL,
  "is_pending"                   BOOLEAN NOT NULL,
  "created_at"                   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  "deleted_at"                   TIMESTAMP WITH TIME ZONE,

  CONSTRAINT "pk_plaid_transactions"                    PRIMARY KEY ("plaid_transaction_id", "account_id"),
  CONSTRAINT "fk_plaid_transactions_account"            FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_plaid_transactions_plaid_bank_account" FOREIGN KEY ("plaid_bank_account_id", "account_id") REFERENCES "plaid_bank_accounts" ("plaid_bank_account_id", "account_id")
);

INSERT INTO "plaid_transactions" (
  "account_id", 
  "plaid_bank_account_id", 
  "plaid_id", 
  "pending_plaid_id",
  "categories", 
  "date", 
  "authorized_date", 
  "name", 
  "merchant_name", 
  "amount", 
  "currency", 
  "is_pending", 
  "created_at", 
  "deleted_at"
)
SELECT
  "transactions"."account_id",
  "bank_accounts"."plaid_bank_account_id",
  "transactions"."plaid_transaction_id" AS "plaid_id",
  "transactions"."pending_plaid_transaction_id" AS "pending_plaid_id",
  "transactions"."original_categories" AS "categories",
  "transactions"."date",
  "transactions"."authorized_date",
  "transactions"."original_name" AS "name",
  "transactions"."original_merchant_name" AS "merchant_name",
  "transactions"."amount",
  "transactions"."currency",
  "transactions"."is_pending",
  "transactions"."created_at",
  "transactions"."deleted_at"
FROM "transactions"
INNER JOIN "bank_accounts" ON "bank_accounts"."account_id" = "transactions"."account_id" AND
                              "bank_accounts"."bank_account_id" = "transactions"."bank_account_id"
WHERE "bank_accounts"."plaid_bank_account_id" IS NOT NULL;

-- Rename the old plaid ID columns on the current transaction table so we can
-- backfill new ones.
ALTER TABLE "transactions" RENAME COLUMN "plaid_transaction_id" TO "old_plaid_transaction_id";
ALTER TABLE "transactions" RENAME COLUMN "pending_plaid_transaction_id" TO "old_pending_plaid_transaction_id";

ALTER TABLE "transactions"
ADD COLUMN "plaid_transaction_id" BIGINT,
ADD COLUMN "pending_plaid_transaction_id" BIGINT;

-- Backfill the plaid column on the transactions table.
UPDATE "transactions"
SET "plaid_transaction_id" = "plaid_transactions"."plaid_transaction_id"
FROM "plaid_transactions"
WHERE "plaid_transactions"."account_id" = "transactions"."account_id" AND
      "plaid_transactions"."plaid_id" = "transactions"."old_plaid_transaction_id";

-- Backfill the pending column on the transactions table.
UPDATE "transactions"
SET "pending_plaid_transaction_id" = "plaid_transactions"."plaid_transaction_id"
FROM "plaid_transactions"
WHERE "plaid_transactions"."account_id" = "transactions"."account_id" AND
      "plaid_transactions"."plaid_id" = "transactions"."old_pending_plaid_transaction_id";

-- Cleanup the plaid specific columns that we no longer need.
ALTER TABLE "transactions"
DROP COLUMN "old_plaid_transaction_id",
DROP COLUMN "old_pending_plaid_transaction_id",
DROP COLUMN "custom_name",
DROP COLUMN "authorized_date",
DROP COLUMN "original_categories",
ADD CONSTRAINT "fk_transactions_plaid_transactions" FOREIGN KEY ("plaid_transaction_id", "account_id") REFERENCES "plaid_transactions" ("plaid_transaction_id", "account_id"),
ADD CONSTRAINT "fk_transactions_plaid_transactions_pending" FOREIGN KEY ("pending_plaid_transaction_id", "account_id") REFERENCES "plaid_transactions" ("plaid_transaction_id", "account_id");

