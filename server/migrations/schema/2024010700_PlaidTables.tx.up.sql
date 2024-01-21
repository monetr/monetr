/*

This is a huge migration to decouple the common models that monetr uses from the plaid data that they are derived from.
This is done so that adding other data sources in the future should be easier, things like manual import, finicity or 
teller.io.

Here is a general list of what the new tables will look like:

## Plaid Syncs

                                              Table "public.plaid_syncs"
    Column     |           Type           | Collation | Nullable |                      Default
---------------+--------------------------+-----------+----------+----------------------------------------------------
 plaid_sync_id | bigint                   |           | not null | nextval('plaid_syncs_plaid_sync_id_seq'::regclass)
 plaid_link_id | bigint                   |           | not null |
 timestamp     | timestamp with time zone |           | not null |
 trigger       | text                     |           | not null |
 cursor        | text                     |           | not null |
 added         | integer                  |           | not null |
 modified      | integer                  |           | not null |
 removed       | integer                  |           | not null |
 account_id    | bigint                   |           | not null |
Indexes:
    "plaid_syncs_pkey" PRIMARY KEY, btree (plaid_sync_id, account_id)
    "ix_plaid_syncs_timestamp" btree (plaid_link_id, "timestamp" DESC)
Foreign-key constraints:
    "fk_plaid_syncs_account" FOREIGN KEY (account_id) REFERENCES accounts(account_id)
    "fk_plaid_syncs_plaid_link" FOREIGN KEY (plaid_link_id, account_id) REFERENCES plaid_links(plaid_link_id, account_id)

## Plaid Links

                                                  Table "public.plaid_links"
         Column         |           Type           | Collation | Nullable |                      Default
------------------------+--------------------------+-----------+----------+----------------------------------------------------
 plaid_link_id          | bigint                   |           | not null | nextval('plaid_links_plaid_link_id_seq'::regclass)
 item_id                | text                     |           | not null |
 products               | text[]                   |           |          |
 webhook_url            | text                     |           |          |
 institution_id         | text                     |           | not null |
 institution_name       | text                     |           | not null |
 account_id             | bigint                   |           | not null |
 status                 | integer                  |           | not null | 0
 error_code             | text                     |           |          |
 expiration_date        | timestamp with time zone |           |          |
 new_accounts_available | boolean                  |           | not null | false
 last_manual_sync       | timestamp with time zone |           |          |
 last_successful_update | timestamp with time zone |           |          |
 last_attempted_update  | timestamp with time zone |           |          |
 updated_at             | timestamp with time zone |           | not null | now()
 created_at             | timestamp with time zone |           | not null | now()
 created_by_user_id     | bigint                   |           | not null |
Indexes:
    "plaid_links_pkey" PRIMARY KEY, btree (plaid_link_id, account_id)
    "uq_plaid_links_item_id" UNIQUE CONSTRAINT, btree (item_id)
Foreign-key constraints:
    "fk_plaid_links_account" FOREIGN KEY (account_id) REFERENCES accounts(account_id)
Referenced by:
    TABLE "links" CONSTRAINT "fk_links_plaid_link" FOREIGN KEY (plaid_link_id, account_id) REFERENCES plaid_links(plaid_link_id, account_id)
    TABLE "plaid_bank_accounts" CONSTRAINT "fk_plaid_bank_accounts_plaid_link" FOREIGN KEY (plaid_link_id, account_id) REFERENCES plaid_links(plaid_link_id, account_id)
    TABLE "plaid_syncs" CONSTRAINT "fk_plaid_syncs_plaid_link" FOREIGN KEY (plaid_link_id, account_id) REFERENCES plaid_links(plaid_link_id, account_id)

## Plaid Bank Accounts

                                                      Table "public.plaid_bank_accounts"
        Column         |           Type           | Collation | Nullable |                              Default
-----------------------+--------------------------+-----------+----------+--------------------------------------------------------------------
 plaid_bank_account_id | bigint                   |           | not null | nextval('plaid_bank_accounts_plaid_bank_account_id_seq'::regclass)
 account_id            | bigint                   |           | not null |
 plaid_link_id         | bigint                   |           | not null |
 plaid_id              | text                     |           | not null |
 name                  | text                     |           | not null |
 official_name         | text                     |           |          |
 mask                  | text                     |           |          |
 available_balance     | bigint                   |           | not null |
 current_balance       | bigint                   |           | not null |
 limit_balance         | bigint                   |           |          |
 created_at            | timestamp with time zone |           | not null | now()
 created_by_user_id    | bigint                   |           | not null |
Indexes:
    "pk_plaid_bank_accounts" PRIMARY KEY, btree (plaid_bank_account_id, account_id)
Foreign-key constraints:
    "fk_plaid_bank_accounts_account" FOREIGN KEY (account_id) REFERENCES accounts(account_id)
    "fk_plaid_bank_accounts_plaid_link" FOREIGN KEY (plaid_link_id, account_id) REFERENCES plaid_links(plaid_link_id, account_id)
    "fk_plaid_bank_accounts_users_created_by_user_id" FOREIGN KEY (created_by_user_id) REFERENCES users(user_id)
Referenced by:
    TABLE "bank_accounts" CONSTRAINT "fk_bank_accounts_plaid_bank_accounts" FOREIGN KEY (plaid_bank_account_id, account_id) REFERENCES plaid_bank_accounts(plaid_bank_account_id, account_id)
    TABLE "plaid_transactions" CONSTRAINT "fk_plaid_transactions_plaid_bank_account" FOREIGN KEY (plaid_bank_account_id, account_id) REFERENCES plaid_bank_accounts(plaid_bank_account_id, account_id)

## Plaid Transactions

                                                     Table "public.plaid_transactions"
        Column         |           Type           | Collation | Nullable |                             Default
-----------------------+--------------------------+-----------+----------+------------------------------------------------------------------
 plaid_transaction_id  | bigint                   |           | not null | nextval('plaid_transactions_plaid_transaction_id_seq'::regclass)
 account_id            | bigint                   |           | not null |
 plaid_bank_account_id | bigint                   |           | not null |
 plaid_id              | text                     |           | not null |
 pending_plaid_id      | text                     |           |          |
 categories            | text[]                   |           |          |
 date                  | timestamp with time zone |           | not null |
 authorized_date       | timestamp with time zone |           |          |
 name                  | text                     |           | not null |
 merchant_name         | text                     |           |          |
 amount                | bigint                   |           | not null |
 currency              | text                     |           | not null |
 is_pending            | boolean                  |           | not null |
 created_at            | timestamp with time zone |           | not null | now()
 deleted_at            | timestamp with time zone |           |          |
Indexes:
    "pk_plaid_transactions" PRIMARY KEY, btree (plaid_transaction_id, account_id)
Foreign-key constraints:
    "fk_plaid_transactions_account" FOREIGN KEY (account_id) REFERENCES accounts(account_id)
    "fk_plaid_transactions_plaid_bank_account" FOREIGN KEY (plaid_bank_account_id, account_id) REFERENCES plaid_bank_accounts(plaid_bank_account_id, account_id)
Referenced by:
    TABLE "transactions" CONSTRAINT "fk_transactions_plaid_transactions" FOREIGN KEY (plaid_transaction_id, account_id) REFERENCES plaid_transactions(plaid_transaction_id, account_id)
    TABLE "transactions" CONSTRAINT "fk_transactions_plaid_transactions_pending" FOREIGN KEY (pending_plaid_transaction_id, account_id) REFERENCES plaid_transactions(plaid_transaction_id, account_id)

# Model Tables

Several model tables have also been changed as part of this migration. Their new schema is listed below:

## Links

                                             Table "public.links"
       Column       |           Type           | Collation | Nullable |                Default
--------------------+--------------------------+-----------+----------+----------------------------------------
 link_id            | bigint                   |           | not null | nextval('links_link_id_seq'::regclass)
 account_id         | bigint                   |           | not null |
 link_type          | smallint                 |           | not null |
 plaid_link_id      | bigint                   |           |          |
 institution_name   | text                     |           |          |
 created_at         | timestamp with time zone |           | not null |
 created_by_user_id | bigint                   |           | not null |
 updated_at         | timestamp with time zone |           | not null |
 description        | text                     |           |          |
Indexes:
    "pk_links" PRIMARY KEY, btree (link_id, account_id)
Foreign-key constraints:
    "fk_links_accounts_account_id" FOREIGN KEY (account_id) REFERENCES accounts(account_id)
    "fk_links_plaid_link" FOREIGN KEY (plaid_link_id, account_id) REFERENCES plaid_links(plaid_link_id, account_id)
    "fk_links_users_created_by_user_id" FOREIGN KEY (created_by_user_id) REFERENCES users(user_id)
Referenced by:
    TABLE "bank_accounts" CONSTRAINT "fk_bank_accounts_links_link_id_account_id" FOREIGN KEY (link_id, account_id) REFERENCES links(link_id, account_id)

## Bank Accounts

                                                   Table "public.bank_accounts"
        Column         |           Type           | Collation | Nullable |                        Default
-----------------------+--------------------------+-----------+----------+--------------------------------------------------------
 bank_account_id       | bigint                   |           | not null | nextval('bank_accounts_bank_account_id_seq'::regclass)
 account_id            | bigint                   |           | not null |
 link_id               | bigint                   |           | not null |
 available_balance     | bigint                   |           | not null |
 current_balance       | bigint                   |           | not null |
 mask                  | text                     |           |          |
 name                  | text                     |           | not null |
 account_type          | text                     |           |          |
 account_sub_type      | text                     |           |          |
 last_updated          | timestamp with time zone |           | not null | (now() AT TIME ZONE 'UTC'::text)
 status                | text                     |           | not null | 'active'::text
 original_name         | text                     |           |          |
 plaid_bank_account_id | bigint                   |           |          |
 updated_at            | timestamp with time zone |           | not null | now()
 created_at            | timestamp with time zone |           | not null | now()
 created_by_user_id    | bigint                   |           | not null |
Indexes:
    "pk_bank_accounts" PRIMARY KEY, btree (bank_account_id, account_id)
Foreign-key constraints:
    "fk_bank_accounts_accounts_account_id" FOREIGN KEY (account_id) REFERENCES accounts(account_id)
    "fk_bank_accounts_links_link_id_account_id" FOREIGN KEY (link_id, account_id) REFERENCES links(link_id, account_id)
    "fk_bank_accounts_plaid_bank_accounts" FOREIGN KEY (plaid_bank_account_id, account_id) REFERENCES plaid_bank_accounts(plaid_bank_account_id, account_id)
Referenced by:
    TABLE "files" CONSTRAINT "fk_files_bank_account" FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts(bank_account_id, account_id)
    TABLE "funding_schedules" CONSTRAINT "fk_funding_schedules_bank_accounts_bank_account_id_account_id" FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts(bank_account_id, account_id)
    TABLE "spending" CONSTRAINT "fk_spending_bank_accounts_bank_account_id_account_id" FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts(bank_account_id, account_id)
    TABLE "transaction_clusters" CONSTRAINT "fk_transaction_clusters_bank_account" FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts(bank_account_id, account_id)
    TABLE "transactions" CONSTRAINT "fk_transactions_bank_accounts_bank_account_id_account_id" FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts(bank_account_id, account_id)

## Transactions
                                                      Table "public.transactions"
            Column            |           Type           | Collation | Nullable |                       Default
------------------------------+--------------------------+-----------+----------+------------------------------------------------------
 transaction_id               | bigint                   |           | not null | nextval('transactions_transaction_id_seq'::regclass)
 account_id                   | bigint                   |           | not null |
 bank_account_id              | bigint                   |           | not null |
 amount                       | bigint                   |           | not null |
 spending_id                  | bigint                   |           |          |
 spending_amount              | bigint                   |           |          |
 categories                   | text[]                   |           |          |
 name                         | text                     |           |          |
 original_name                | text                     |           | not null |
 merchant_name                | text                     |           |          |
 original_merchant_name       | text                     |           |          |
 is_pending                   | boolean                  |           | not null |
 created_at                   | timestamp with time zone |           | not null | now()
 date                         | timestamp with time zone |           | not null |
 deleted_at                   | timestamp with time zone |           |          |
 currency                     | text                     |           | not null | 'USD'::text
 plaid_transaction_id         | bigint                   |           |          |
 pending_plaid_transaction_id | bigint                   |           |          |
Indexes:
    "pk_transactions" PRIMARY KEY, btree (transaction_id, account_id, bank_account_id)
    "ix_transactions_opt_order" btree (account_id, bank_account_id, date DESC, transaction_id DESC)
    "ix_transactions_soft_delete" btree (account_id, bank_account_id, date DESC, transaction_id DESC) WHERE deleted_at IS NULL
Foreign-key constraints:
    "fk_transactions_accounts_account_id" FOREIGN KEY (account_id) REFERENCES accounts(account_id)
    "fk_transactions_bank_accounts_bank_account_id_account_id" FOREIGN KEY (bank_account_id, account_id) REFERENCES bank_accounts(bank_account_id, account_id)
    "fk_transactions_plaid_transactions" FOREIGN KEY (plaid_transaction_id, account_id) REFERENCES plaid_transactions(plaid_transaction_id, account_id)
    "fk_transactions_plaid_transactions_pending" FOREIGN KEY (pending_plaid_transaction_id, account_id) REFERENCES plaid_transactions(plaid_transaction_id, account_id)
    "fk_transactions_spending" FOREIGN KEY (spending_id, account_id, bank_account_id) REFERENCES spending(spending_id, account_id, bank_account_id) ON DELETE SET NULL

*/



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

