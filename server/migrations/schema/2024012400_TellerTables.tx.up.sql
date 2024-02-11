CREATE TABLE "secrets" (
  "secret_id"          BIGSERIAL NOT NULL,
  "account_id"         BIGINT    NOT NULL,
  "kind"               TEXT      NOT NULL,
  "key_id"             TEXT,
  "version"            TEXT,
  "secret"             TEXT      NOT NULL,
  "updated_at"         TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_at"         TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  CONSTRAINT "pk_secrets"                          PRIMARY KEY ("secret_id", "account_id"),
  CONSTRAINT "fk_secrets_account"                  FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id")
);

ALTER TABLE "secrets" ADD COLUMN "plaid_link_id" BIGINT NOT NULL;
INSERT INTO "secrets" ("account_id", "plaid_link_id", "kind", "key_id", "version", "secret", "updated_at", "created_at")
SELECT
  "plaid_token"."account_id",
  "plaid_link"."plaid_link_id",
  'plaid' AS "kind",
  "plaid_token"."key_id",
  "plaid_token"."version",
  "plaid_token"."access_token" AS "secret",
  "plaid_link"."updated_at" AS "updated_at",
  "plaid_link"."created_at" AS "created_at"
FROM "plaid_tokens" AS "plaid_token"
INNER JOIN "plaid_links" AS "plaid_link" ON "plaid_link"."account_id" = "plaid_token"."account_id" AND "plaid_token"."item_id" = "plaid_link"."item_id";

ALTER TABLE "plaid_links" ADD COLUMN "secret_id" BIGINT;

UPDATE "plaid_links"
SET "secret_id" = "secrets"."secret_id"
FROM "secrets"
WHERE "secrets"."account_id" = "plaid_links"."account_id" AND "secrets"."plaid_link_id" =  "plaid_links"."plaid_link_id";

ALTER TABLE "plaid_links" ALTER COLUMN "secret_id" SET NOT NULL;

ALTER TABLE "secrets" 
DROP COLUMN "plaid_link_id";

ALTER TABLE "plaid_links" 
ADD CONSTRAINT "fk_plaid_links_secret" FOREIGN KEY ("secret_id", "account_id") REFERENCES "secrets" ("secret_id", "account_id");

DROP TABLE "plaid_tokens";

CREATE TABLE "teller_links" (
  "teller_link_id"         BIGSERIAL NOT NULL,
  "account_id"             BIGINT    NOT NULL,
  "enrollment_id"          TEXT      NOT NULL,
  "teller_user_id"         TEXT      NOT NULL,
  "status"                 INT       NOT NULL,
  "error_code"             TEXT,
  "institution_name"       TEXT      NOT NULL,
  "last_manual_sync"       TIMESTAMP WITH TIME ZONE,
  "last_successful_update" TIMESTAMP WITH TIME ZONE,
  "last_attempted_update"  TIMESTAMP WITH TIME ZONE,
  "secret_id"              BIGINT NOT NULL,
  "updated_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_by_user_id"     BIGINT    NOT NULL,
  CONSTRAINT "pk_teller_links"                          PRIMARY KEY ("teller_link_id", "account_id"),
  CONSTRAINT "fk_teller_links_account"                  FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_teller_links_secret"                   FOREIGN KEY ("secret_id", "account_id") REFERENCES "secrets" ("secret_id", "account_id"),
  CONSTRAINT "fk_teller_links_users_created_by_user_id" FOREIGN KEY ("created_by_user_id") REFERENCES "users" ("user_id"),
  CONSTRAINT "uq_teller_links_enrollment"               UNIQUE ("account_id", "enrollment_id")
);

ALTER TABLE "links" 
ADD COLUMN "teller_link_id" BIGINT,
ADD COLUMN "deleted_at" TIMESTAMP WITH TIME ZONE,
ADD CONSTRAINT "fk_links_teller_link" FOREIGN KEY ("teller_link_id", "account_id") REFERENCES "teller_links" ("teller_link_id", "account_id");

ALTER TABLE "bank_accounts"
DROP COLUMN "created_by_user_id";

CREATE TABLE "teller_bank_accounts" (
  "teller_bank_account_id" BIGSERIAL NOT NULL,
  "account_id"             BIGINT    NOT NULL,
  "teller_link_id"         BIGINT    NOT NULL,
  "teller_id"              TEXT      NOT NULL,
  "institution_id"         TEXT      NOT NULL,
  "institution_name"       TEXT      NOT NULL,
  "mask"                   TEXT      NOT NULL,
  "name"                   TEXT      NOT NULL,
  "type"                   TEXT      NOT NULL,
  "sub_type"               TEXT      NOT NULL,
  "status"                 INT       NOT NULL,
  "ledger_balance"         BIGINT,
  "updated_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "created_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "balanced_at"            TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_teller_bank_accounts"                          PRIMARY KEY ("teller_bank_account_id", "account_id"),
  CONSTRAINT "fk_teller_bank_accounts_account"                  FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_teller_bank_accounts_teller_link"              FOREIGN KEY ("teller_link_id", "account_id") REFERENCES "teller_links" ("teller_link_id", "account_id"),
  CONSTRAINT "uq_teller_bank_accounts_teller_id"                UNIQUE ("account_id", "teller_id")
);

ALTER TABLE "bank_accounts"
ADD COLUMN "teller_bank_account_id" BIGINT,
ADD CONSTRAINT "fk_bank_accounts_teller_bank_account" FOREIGN KEY ("teller_bank_account_id", "account_id") REFERENCES "teller_bank_accounts" ("teller_bank_account_id", "account_id");

CREATE TABLE "teller_syncs" (
  "teller_sync_id"         BIGSERIAL NOT NULL,
  "account_id"             BIGINT    NOT NULL,
  "teller_bank_account_id" BIGINT    NOT NULL,
  "timestamp"              TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "trigger"                TEXT NOT NULL,
  "immutable_timestamp"    TIMESTAMP WITH TIME ZONE NOT NULL,
  "added"                  INT NOT NULL,
  "modified"               INT NOT NULL,
  "removed"                INT NOT NULL,
  CONSTRAINT "pk_teller_syncs"                     PRIMARY KEY ("teller_sync_id", "account_id"),
  CONSTRAINT "fk_teller_syncs_account"             FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_teller_syncs_teller_bank_account" FOREIGN KEY ("teller_bank_account_id", "account_id") REFERENCES "teller_bank_accounts" ("teller_bank_account_id", "account_id")
);

CREATE TABLE "teller_transactions" (
  "teller_transaction_id"  BIGSERIAL NOT NULL,
  "account_id"             BIGINT    NOT NULL,
  "teller_bank_account_id" BIGINT    NOT NULL,
  "teller_id"              TEXT      NOT NULL,
  "name"                   TEXT      NOT NULL,
  "category"               TEXT,
  "type"                   TEXT,
  "date"                   TIMESTAMP WITH TIME ZONE NOT NULL,
  "is_pending"             BOOL   NOT NULL,
  "amount"                 BIGINT NOT NULL,
  "running_balance"        BIGINT,
  "created_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "updated_at"             TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
  "deleted_at"             TIMESTAMP WITH TIME ZONE,
  CONSTRAINT "pk_teller_transactions"                     PRIMARY KEY ("teller_transaction_id", "account_id"),
  CONSTRAINT "fk_teller_transactions_account"             FOREIGN KEY ("account_id") REFERENCES "accounts" ("account_id"),
  CONSTRAINT "fk_teller_transactions_teller_bank_account" FOREIGN KEY ("teller_bank_account_id", "account_id") REFERENCES "teller_bank_accounts" ("teller_bank_account_id", "account_id"),
  CONSTRAINT "uq_teller_transactions_teller_id"           UNIQUE ("teller_bank_account_id", "teller_id")
);

ALTER TABLE "transactions"
ADD COLUMN "teller_transaction_id" BIGINT,
ADD CONSTRAINT "fk_transactions_teller_transaction" FOREIGN KEY ("teller_transaction_id", "account_id") REFERENCES "teller_transactions" ("teller_transaction_id", "account_id");

-- We are going to change the entire table, so we need to clean out the table as part of this.
DELETE FROM "jobs" WHERE 1=1;
ALTER TABLE "jobs" 
DROP COLUMN "input",
DROP COLUMN "output",
ADD COLUMN "input" JSONB,
ADD COLUMN "output" JSONB;
