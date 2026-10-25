CREATE TABLE "simplefin_links" (
  "simplefin_link_id"      VARCHAR(32) NOT NULL,
  "account_id"             VARCHAR(32) NOT NULL,
  "secret_id"              VARCHAR(32) NOT NULL,
  "name"                   TEXT        NOT NULL,
  "api_url"                TEXT        NOT NULL,
  "status"                 TEXT        NOT NULL,
  "last_errors"            JSONB,
  "last_manual_sync"       TIMESTAMP WITHOUT TIME ZONE,
  "last_successful_update" TIMESTAMP WITHOUT TIME ZONE,
  "last_attempted_update"  TIMESTAMP WITHOUT TIME ZONE,
  "created_by"             VARCHAR(32) NOT NULL,
  "created_at"             TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "updated_at"             TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at"             TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_simplefin_links"            PRIMARY KEY ("simplefin_link_id", "account_id"),
  CONSTRAINT "fk_simplefin_links_account"    FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_simplefin_links_secret"     FOREIGN KEY ("secret_id", "account_id") REFERENCES "secrets" ("secret_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_simplefin_links_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("user_id") ON DELETE CASCADE
);

ALTER TABLE "links" ADD COLUMN "simplefin_link_id" VARCHAR(32);

-- Same as lunch flow, restrict deletes because if we fail to update the link
-- type to be manual then the link is corrupt.
ALTER TABLE "links" ADD CONSTRAINT "fk_links_simplefin_link" FOREIGN KEY ("simplefin_link_id", "account_id")
REFERENCES "simplefin_links" ("simplefin_link_id", "account_id") ON DELETE RESTRICT;

ALTER TABLE "links" ADD CONSTRAINT "uq_links_simplefin_link" UNIQUE ("simplefin_link_id");

CREATE TABLE "simplefin_bank_accounts" (
  "simplefin_bank_account_id" VARCHAR(32) NOT NULL,
  "account_id"                VARCHAR(32) NOT NULL,
  "simplefin_link_id"         VARCHAR(32) NOT NULL,
  "simplefin_id"              TEXT        NOT NULL,
  "connection_id"             TEXT        NOT NULL,
  "name"                      TEXT        NOT NULL,
  "institution_name"          TEXT        NOT NULL,
  "currency"                  TEXT        NOT NULL,
  "status"                    TEXT        NOT NULL,
  "current_balance"           BIGINT      NOT NULL,
  "available_balance"         BIGINT      NOT NULL,
  "created_by"                VARCHAR(32) NOT NULL,
  "created_at"                TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "updated_at"                TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at"                TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_simplefin_bank_accounts"                PRIMARY KEY ("simplefin_bank_account_id", "account_id"),
  CONSTRAINT "fk_simplefin_bank_accounts_account"        FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_simplefin_bank_accounts_simplefin_link" FOREIGN KEY ("simplefin_link_id", "account_id") REFERENCES "simplefin_links" ("simplefin_link_id", "account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_simplefin_bank_accounts_created_by"     FOREIGN KEY ("created_by") REFERENCES "users" ("user_id"),
  -- SimpleFIN only promises that an account id is unique within its connection,
  -- not across the whole link. So the connection has to be part of this too.
  -- See: https://www.simplefin.org/protocol.html#account
  CONSTRAINT "uq_simplefin_bank_accounts_simplefin_id"   UNIQUE ("simplefin_link_id", "connection_id", "simplefin_id")
);

ALTER TABLE "bank_accounts" ADD COLUMN "simplefin_bank_account_id" VARCHAR(32);
ALTER TABLE "bank_accounts" ADD CONSTRAINT "fk_bank_accounts_simplefin_bank_account" FOREIGN KEY ("simplefin_bank_account_id", "account_id") REFERENCES "simplefin_bank_accounts" ("simplefin_bank_account_id", "account_id") ON DELETE CASCADE;
ALTER TABLE "bank_accounts" ADD CONSTRAINT "uq_bank_accounts_simplefin_bank_account" UNIQUE ("simplefin_bank_account_id");

CREATE TABLE "simplefin_transactions" (
  "simplefin_transaction_id"  VARCHAR(32) NOT NULL,
  "account_id"                VARCHAR(32) NOT NULL,
  "simplefin_bank_account_id" VARCHAR(32) NOT NULL,
  "simplefin_id"              TEXT        NOT NULL,
  "description"               TEXT        NOT NULL,
  "amount"                    BIGINT      NOT NULL,
  "posted"                    TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "transacted_at"             TIMESTAMP WITHOUT TIME ZONE,
  "created_at"                TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  "deleted_at"                TIMESTAMP WITHOUT TIME ZONE,
  CONSTRAINT "pk_simplefin_transactions"                        PRIMARY KEY ("simplefin_transaction_id", "account_id"),
  CONSTRAINT "fk_simplefin_transactions_account"                FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id") ON DELETE CASCADE,
  CONSTRAINT "fk_simplefin_transactions_simplefin_bank_account" FOREIGN KEY ("simplefin_bank_account_id", "account_id") REFERENCES "simplefin_bank_accounts" ("simplefin_bank_account_id", "account_id"),
  CONSTRAINT "uq_simplefin_transactions_simplefin_id"           UNIQUE ("simplefin_bank_account_id", "simplefin_id")
);

ALTER TABLE "transactions" ADD COLUMN "simplefin_transaction_id" VARCHAR(32);
ALTER TABLE "transactions" ADD CONSTRAINT "fk_transactions_simplefin_transaction" FOREIGN KEY ("simplefin_transaction_id", "account_id") REFERENCES "simplefin_transactions" ("simplefin_transaction_id", "account_id") ON DELETE CASCADE;
ALTER TABLE "transactions" ADD CONSTRAINT "uq_transactions_simplefin_transaction" UNIQUE ("simplefin_transaction_id");
