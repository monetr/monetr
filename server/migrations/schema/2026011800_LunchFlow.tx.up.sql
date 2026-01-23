CREATE TABLE "lunch_flow_links" (
  "lunch_flow_link_id"     VARCHAR(32) NOT NULL,
  "account_id"             VARCHAR(32) NOT NULL,
  "secret_id"              VARCHAR(32) NOT NULL,
  "api_url"                TEXT        NOT NULL,
  "status"                 TEXT        NOT NULL,
  "last_manual_sync"       TIMESTAMP WITHOUT TIME ZONE,
  "last_successful_update" TIMESTAMP WITHOUT TIME ZONE,
  "last_attempted_update"  TIMESTAMP WITHOUT TIME ZONE,
  "created_by"             VARCHAR(32) NOT NULL,
  "updated_at"             TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "created_at"             TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at"             TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_lunch_flow_links"            PRIMARY KEY ("lunch_flow_link_id", "account_id"),
  CONSTRAINT "fk_lunch_flow_links_account"    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_lunch_flow_links_secret"     FOREIGN KEY ("secret_id", "account_id") REFERENCES "secrets" ("secret_id", "account_id"),
  CONSTRAINT "fk_lunch_flow_links_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id")
);

ALTER TABLE "links" ADD COLUMN "lunch_flow_link_id" VARCHAR(32);
ALTER TABLE "links" ADD CONSTRAINT "fk_links_lunch_flow_link" FOREIGN KEY ("lunch_flow_link_id", "account_id") REFERENCES "lunch_flow_links" ("lunch_flow_link_id", "account_id") ON DELETE SET NULL;
-- You cannot have the same lunch flow link associated with multiple links. You
-- must create multiple lunch flow links!
ALTER TABLE "links" ADD CONSTRAINT "uq_links_lunch_flow_link" UNIQUE ("lunch_flow_link_id");

CREATE TABLE "lunch_flow_bank_accounts" (
  "lunch_flow_bank_account_id" VARCHAR(32) NOT NULL,
  "lunch_flow_link_id"         VARCHAR(32) NOT NULL,
  "account_id"                 VARCHAR(32) NOT NULL,
  "lunch_flow_id"              TEXT        NOT NULL,
  "lunch_flow_status"          TEXT        NOT NULL,
  "name"                       TEXT        NOT NULL,
  "institution_name"           TEXT        NOT NULL,
  "provider"                   TEXT        NOT NULL,
  "currency"                   TEXT        NOT NULL,
  "status"                     TEXT        NOT NULL,
  "current_balance"            BIGINT      NOT NULL,
  "created_by"                 VARCHAR(32) NOT NULL,
  "updated_at"                 TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "created_at"                 TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at"                 TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_lunch_flow_bank_accounts"                 PRIMARY KEY ("lunch_flow_bank_account_id", "account_id"),
  CONSTRAINT "fk_lunch_flow_bank_accounts_account"         FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_lunch_flow_bank_accounts_lunch_flow_link" FOREIGN KEY ("lunch_flow_link_id", "account_id") REFERENCES "lunch_flow_links" ("lunch_flow_link_id", "account_id"),
  CONSTRAINT "fk_lunch_flow_bank_accounts_created_by"      FOREIGN KEY ("created_by") REFERENCES "users" ("user_id"),
  -- Maintain uniqueness per link, we cannot have an account twice within the
  -- same link. However if the user wants to have the same data multiple times
  -- for some reason, they can create an additional link with the same API URL
  -- and secret and that will work.
  CONSTRAINT "uq_lunch_flow_bank_accounts_lunch_flow_id"   UNIQUE ("lunch_flow_link_id", "lunch_flow_id")
);

ALTER TABLE "bank_accounts" ADD COLUMN "lunch_flow_bank_account_id" VARCHAR(32);
ALTER TABLE "bank_accounts" ADD CONSTRAINT "fk_bank_accounts_lunch_flow_bank_account" FOREIGN KEY ("lunch_flow_bank_account_id", "account_id") REFERENCES "lunch_flow_bank_accounts" ("lunch_flow_bank_account_id", "account_id") ON DELETE SET NULL;
ALTER TABLE "bank_accounts" ADD CONSTRAINT "uq_bank_accounts_lunch_flow_bank_account" UNIQUE ("lunch_flow_bank_account_id");

CREATE TABLE "lunch_flow_transactions" (
  "lunch_flow_transaction_id"  VARCHAR(32) NOT NULL,
  "account_id"                 VARCHAR(32) NOT NULL,
  "lunch_flow_bank_account_id" VARCHAR(32) NOT NULL,
  "lunch_flow_id"              TEXT        NOT NULL,
  "merchant"                   TEXT        NOT NULL,
  "description"                TEXT        NOT NULL,
  "currency"                   TEXT        NOT NULL,
  "amount"                     BIGINT      NOT NULL,
  "is_pending"                 BOOLEAN     NOT NULL DEFAULT false,
  "date"                       TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "created_at"                 TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at"                 TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_lunch_flow_transactions"                         PRIMARY KEY ("lunch_flow_transaction_id", "account_id"),
  CONSTRAINT "fk_lunch_flow_transactions_account"                 FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_lunch_flow_transactions_lunch_flow_bank_account" FOREIGN KEY ("lunch_flow_bank_account_id", "account_id") REFERENCES "lunch_flow_bank_accounts" ("lunch_flow_bank_account_id", "account_id"),
  -- Only allow a transaction to appear once per lunch flow bank account. Again
  -- if the user wants to have data duplicated for some reason they need to do
  -- that on a link level.
  CONSTRAINT "uq_lunch_flow_transactions_lunch_flow_id"           UNIQUE ("lunch_flow_bank_account_id", "lunch_flow_id")
);

ALTER TABLE "transactions" ADD COLUMN "lunch_flow_transaction_id" VARCHAR(32);
ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_lunch_flow_transaction" FOREIGN KEY ("lunch_flow_transaction_id", "account_id") REFERENCES "lunch_flow_transactions" ("lunch_flow_transaction_id", "account_id") ON DELETE SET NULL;
ALTER TABLE "transactions" ADD CONSTRAINT "uq_transactions_lunch_flow_bank_account" UNIQUE ("lunch_flow_transaction_id");

-- Update the enum for the link type too.
ALTER TABLE "links" ADD COLUMN "link_type_new" TEXT NOT NULL DEFAULT 'unknown';
UPDATE "links" SET "link_type_new" = 'plaid' WHERE "link_type" = 1;
UPDATE "links" SET "link_type_new" = 'manual' WHERE "link_type" = 2;
UPDATE "links" SET "link_type_new" = 'stripe' WHERE "link_type" = 3;
UPDATE "links" SET "link_type_new" = 'lunch_flow' WHERE "link_type" = 4;
ALTER TABLE "links" DROP COLUMN "link_type";
ALTER TABLE "links" RENAME COLUMN "link_type_new" TO "link_type";
